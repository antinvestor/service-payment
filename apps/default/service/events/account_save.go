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

package events

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/util"
)

type AccountSave struct {
	accountRepo repository.AccountRepository
}

// NewAccountSave creates a new AccountSave event handler with the required dependencies.
func NewAccountSave(accountRepo repository.AccountRepository) *AccountSave {
	return &AccountSave{
		accountRepo: accountRepo,
	}
}

func (e *AccountSave) Name() string {
	return EventNameAccountSave
}

func (e *AccountSave) PayloadType() any {
	return &models.Account{}
}

func (e *AccountSave) Validate(_ context.Context, payload any) error {
	account, ok := payload.(*models.Account)
	if !ok {
		return errors.New("payload is not of type models.Account")
	}

	if account.ID == "" {
		return errors.New("account ID should already have been set")
	}

	return nil
}

func (e *AccountSave) Execute(ctx context.Context, payload any) error {
	account, ok := payload.(*models.Account)
	if !ok {
		return errors.New("payload is not of type models.Account")
	}

	logger := util.Log(ctx).WithFields(map[string]any{"account_id": account.ID, "type": e.Name()})
	logger.Debug("handling event")

	if account.Version > 0 {
		return nil
	}

	err := e.accountRepo.Create(ctx, account)
	if err != nil {
		if data.ErrorIsDuplicateKey(err) {
			logger.Debug("record already exists, skipping duplicate")
			return nil
		}
		logger.WithError(err).Error("could not save account to db")
		return err
	}

	logger.Debug("successfully saved record to db")
	return nil
}
