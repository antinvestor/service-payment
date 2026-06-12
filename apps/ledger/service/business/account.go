// Copyright 2023-2026 Ant Investor Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package business

import (
	"context"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/apps/ledger/service/observability"
	"github.com/antinvestor/service-payments/apps/ledger/service/repository"
	"github.com/antinvestor/service-payments/pkg/apperrors"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/workerpool"
	"github.com/pitabwire/util/decimalx"
	"golang.org/x/text/currency"
)

// AccountBusiness defines the business interface for account operations.
type AccountBusiness interface {
	CreateAccount(ctx context.Context, req *ledgerv1.CreateAccountRequest) (*ledgerv1.Account, error)
	SearchAccounts(ctx context.Context, req *commonv1.SearchRequest,
		consumer func(ctx context.Context, batch []*ledgerv1.Account) error) error
	GetAccount(ctx context.Context, id string) (*ledgerv1.Account, error)
	UpdateAccount(ctx context.Context, req *ledgerv1.UpdateAccountRequest) (*ledgerv1.Account, error)
	DeleteAccount(ctx context.Context, id string) error
}

// accountBusiness implements the AccountBusiness interface.
type accountBusiness struct {
	workMan     workerpool.Manager
	accountRepo repository.AccountRepository
	ledgerRepo  repository.LedgerRepository
	metrics     *observability.Metrics
}

// NewAccountBusiness creates a new account business instance.
func NewAccountBusiness(
	workMan workerpool.Manager,
	ledgerRepo repository.LedgerRepository,
	accountRepo repository.AccountRepository,
) AccountBusiness {
	return &accountBusiness{
		workMan:     workMan,
		accountRepo: accountRepo,
		ledgerRepo:  ledgerRepo,
		metrics:     observability.NewMetrics(),
	}
}

// CreateAccount creates a new account with business validation.
func (b *accountBusiness) CreateAccount(
	ctx context.Context,
	req *ledgerv1.CreateAccountRequest,
) (*ledgerv1.Account, error) {
	ctx, span := b.metrics.StartSpan(ctx, "CreateAccount")
	var err error
	defer func() { b.metrics.EndSpan(ctx, span, err) }()

	// Validate and normalize currency
	currencyUnit, parseErr := currency.ParseISO(req.GetCurrency())
	if parseErr != nil {
		err = ErrAccountCurrencyInvalid
		return nil, err
	}

	if req.GetLedgerId() == "" {
		err = ErrAccountLedgerIDRequired
		return nil, err
	}

	var ledger *models.Ledger
	ledger, err = b.ledgerRepo.GetByID(ctx, req.GetLedgerId())
	if err != nil {
		return nil, err
	}

	zero := decimalx.Zero()
	// Convention by default: AccountType + NormalBalance derive from the
	// parent ledger's type. Callers that need contra / clearing / suspense
	// / memo accounts will supply explicit overrides on a future API path.
	accountType := models.AccountTypeFromLedgerType(ledger.Type)
	normalBalance := models.NormalBalanceForAccountType(accountType)

	accountModel := &models.Account{
		LedgerID:      ledger.GetID(),
		LedgerType:    ledger.Type,
		AccountType:   accountType,
		NormalBalance: normalBalance,
		BookID:        ledger.BookID,
		Currency:      req.GetCurrency(),
		Balance:       zero.Ptr(),
		Data:          req.GetData().AsMap(),
	}

	accountModel.GenID(ctx)
	if req.GetId() != "" {
		accountModel.ID = req.GetId()
	}

	accountModel.Currency = currencyUnit.String()
	// Create the account through repository
	err = b.accountRepo.Create(ctx, accountModel)
	if err != nil {
		return nil, err
	}

	b.metrics.RecordAccountCreated(ctx)

	// Convert back to API type
	return accountModel.ToAPI(), nil
}

// SearchAccounts searches for accounts based on query.
func (b *accountBusiness) SearchAccounts(
	ctx context.Context,
	req *commonv1.SearchRequest, consumer func(ctx context.Context, batch []*ledgerv1.Account) error,
) error {
	// Business logic for search validation
	query := req.GetQuery()
	if query == "" {
		query = "{}" // Default empty query
	}

	// Search through repository
	result, err := b.accountRepo.SearchAsESQ(ctx, query)
	if err != nil {
		return err
	}

	for {
		res, ok := result.ReadResult(ctx)
		if !ok {
			return nil
		}

		if res.IsError() {
			return res.Error()
		}

		var apiResults []*ledgerv1.Account
		for _, account := range res.Item() {
			apiResults = append(apiResults, account.ToAPI())
		}

		jobErr := consumer(ctx, apiResults)
		if jobErr != nil {
			return jobErr
		}
	}
}

// GetAccount retrieves an account by ID.
func (b *accountBusiness) GetAccount(ctx context.Context, id string) (*ledgerv1.Account, error) {
	if id == "" {
		return nil, ErrAccountIDRequired
	}

	account, err := b.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if account == nil {
		return nil, ErrAccountNotFound
	}

	// Convert to API type
	return account.ToAPI(), nil
}

// UpdateAccount updates an existing account.
func (b *accountBusiness) UpdateAccount(
	ctx context.Context,
	req *ledgerv1.UpdateAccountRequest,
) (*ledgerv1.Account, error) {
	// Business logic validation
	if req.GetId() == "" {
		return nil, ErrAccountIDRequired
	}

	// Convert API request to model - need to get existing account first
	existingAccount, err := b.accountRepo.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	if existingAccount == nil {
		return nil, apperrors.ErrAccountNotFound
	}

	// Update fields from request
	if req.GetData() != nil {
		dataMap := &data.JSONMap{}
		existingAccount.Data = dataMap.FromProtoStruct(req.GetData())
	}

	// Update through repository
	_, err = b.accountRepo.Update(ctx, existingAccount)
	if err != nil {
		return nil, err
	}

	// Convert to API type
	return existingAccount.ToAPI(), nil
}

// DeleteAccount deletes an account if it has no transaction entries.
func (b *accountBusiness) DeleteAccount(ctx context.Context, id string) error {
	if id == "" {
		return ErrAccountIDRequired
	}

	hasEntries, err := b.accountRepo.HasTransactionEntries(ctx, id)
	if err != nil {
		return err
	}

	if hasEntries {
		return apperrors.ErrBadDataSupplied.Extend("account has transactions and cannot be deleted")
	}

	return b.accountRepo.Delete(ctx, id)
}
