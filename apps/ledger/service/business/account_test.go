package business_test

import (
	"context"
	"errors"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/ledger/v1"
	"github.com/antinvestor/service-payments/apps/ledger/service/models"
	"github.com/antinvestor/service-payments/apps/ledger/tests"
	_ "github.com/lib/pq"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/frame/frametests/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/structpb"
)

type AccountBusinessSuite struct {
	tests.BaseTestSuite
	ledger *models.Ledger
}

func TestAccountBusinessSuite(t *testing.T) {
	suite.Run(t, new(AccountBusinessSuite))
}

func (as *AccountBusinessSuite) setupFixtures(ctx context.Context, resources *tests.ServiceResources) {
	// Create test ledgers using business layer
	ledgerBusiness := resources.LedgerBusiness

	createLedgerReq := &ledgerv1.CreateLedgerRequest{
		Id:   "test-ledger",
		Type: ledgerv1.LedgerType_ASSET,
	}

	ledger, err := ledgerBusiness.CreateLedger(ctx, createLedgerReq)
	as.Require().NoError(err, "Unable to create ledger for account")

	// Convert to model for test use
	as.ledger = &models.Ledger{
		BaseModel: data.BaseModel{ID: ledger.GetId()},
		Type:      models.FromLedgerType(ledger.GetType()),
	}
}

func (as *AccountBusinessSuite) TestCreateAccountWithBusinessValidation() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)
		as.setupFixtures(ctx, resources)

		// Test creating account through business layer
		accountBusiness := resources.AccountBusiness

		createAccountReq := &ledgerv1.CreateAccountRequest{
			Id:       "test-account",
			LedgerId: as.ledger.ID,
			Currency: "USD",
		}

		account, err := accountBusiness.CreateAccount(ctx, createAccountReq)
		require.NoError(t, err, "Error creating account through business layer")
		require.NotNil(t, account, "Account should be created")

		assert.Equal(t, "test-account", account.GetId(), "Invalid account ID")
		assert.Equal(t, as.ledger.ID, account.GetLedger(), "Invalid ledger ID")
		assert.Equal(t, "USD", account.GetBalance().GetCurrencyCode(), "Invalid currency")
		assert.Equal(t, int64(0), account.GetBalance().GetUnits(), "Initial balance should be zero")
		assert.Equal(t, int32(0), account.GetBalance().GetNanos(), "Initial nanos should be zero")
	})
}

func (as *AccountBusinessSuite) TestCreateAccountWithInvalidCurrency() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)
		as.setupFixtures(ctx, resources)

		accountBusiness := resources.AccountBusiness

		createAccountReq := &ledgerv1.CreateAccountRequest{
			Id:       "invalid-account",
			LedgerId: as.ledger.ID,
			Currency: "INVALID", // Invalid currency code
		}

		account, err := accountBusiness.CreateAccount(ctx, createAccountReq)
		require.Error(t, err, "Should fail with invalid currency")
		assert.Nil(t, account, "Account should not be created")
		assert.Contains(t, err.Error(), "currency is invalid", "Error should mention currency validation")
	})
}

func (as *AccountBusinessSuite) TestCreateAccountWithMissingLedger() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)

		accountBusiness := resources.AccountBusiness

		createAccountReq := &ledgerv1.CreateAccountRequest{
			Id:       "orphan-account",
			LedgerId: "non-existent-ledger",
			Currency: "USD",
		}

		account, err := accountBusiness.CreateAccount(ctx, createAccountReq)
		require.Error(t, err, "Should fail with non-existent ledger")
		assert.Nil(t, account, "Account should not be created")
	})
}

func (as *AccountBusinessSuite) TestGetAccount() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)
		as.setupFixtures(ctx, resources)

		accountBusiness := resources.AccountBusiness

		// First create an account
		createAccountReq := &ledgerv1.CreateAccountRequest{
			Id:       "get-test-account",
			LedgerId: as.ledger.ID,
			Currency: "EUR",
		}

		createdAccount, err := accountBusiness.CreateAccount(ctx, createAccountReq)
		require.NoError(t, err, "Error creating account")

		// Now retrieve it
		retrievedAccount, err := accountBusiness.GetAccount(ctx, "get-test-account")
		require.NoError(t, err, "Error retrieving account")

		assert.Equal(
			t,
			createdAccount.GetId(),
			retrievedAccount.GetId(),
			"Retrieved account should match created account",
		)
		assert.Equal(t, createdAccount.GetLedger(), retrievedAccount.GetLedger(), "Ledger ID should match")
		assert.Equal(
			t,
			createdAccount.GetBalance().GetCurrencyCode(),
			retrievedAccount.GetBalance().GetCurrencyCode(),
			"Currency should match",
		)
	})
}

func (as *AccountBusinessSuite) TestGetAccountNotFound() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)

		accountBusiness := resources.AccountBusiness

		account, err := accountBusiness.GetAccount(ctx, "non-existent-account")
		require.Error(t, err, "Should fail with non-existent account")
		assert.Nil(t, account, "Account should be nil")
	})
}

func (as *AccountBusinessSuite) TestSearchAccounts() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)
		as.setupFixtures(ctx, resources)

		accountBusiness := resources.AccountBusiness

		// Create a couple accounts
		for _, id := range []string{"search-acc-1", "search-acc-2"} {
			_, err := accountBusiness.CreateAccount(ctx, &ledgerv1.CreateAccountRequest{
				Id:       id,
				LedgerId: as.ledger.ID,
				Currency: "USD",
			})
			require.NoError(t, err)
		}

		// Search with empty query
		searchReq := &commonv1.SearchRequest{Query: "{}"}
		var foundAccounts []*ledgerv1.Account
		err := accountBusiness.SearchAccounts(ctx, searchReq, func(_ context.Context, batch []*ledgerv1.Account) error {
			foundAccounts = append(foundAccounts, batch...)
			return nil
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(foundAccounts), 2, "Should find at least 2 accounts")
	})
}

func (as *AccountBusinessSuite) TestSearchAccountsWithQuery() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)
		as.setupFixtures(ctx, resources)

		accountBusiness := resources.AccountBusiness

		_, err := accountBusiness.CreateAccount(ctx, &ledgerv1.CreateAccountRequest{
			Id:       "filter-acc-1",
			LedgerId: as.ledger.ID,
			Currency: "USD",
		})
		require.NoError(t, err)

		searchReq := &commonv1.SearchRequest{
			Query: `{"query": {"must": {"fields": [{"id": {"eq": "filter-acc-1"}}]}}}`,
		}
		var foundAccounts []*ledgerv1.Account
		err = accountBusiness.SearchAccounts(ctx, searchReq, func(_ context.Context, batch []*ledgerv1.Account) error {
			foundAccounts = append(foundAccounts, batch...)
			return nil
		})
		require.NoError(t, err)
		assert.Len(t, foundAccounts, 1)
		assert.Equal(t, "filter-acc-1", foundAccounts[0].GetId())
	})
}

func (as *AccountBusinessSuite) TestDeleteAccountNoTransactions() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)
		as.setupFixtures(ctx, resources)

		accountBusiness := resources.AccountBusiness

		// Create an account with no transactions
		_, err := accountBusiness.CreateAccount(ctx, &ledgerv1.CreateAccountRequest{
			Id:       "delete-me-account",
			LedgerId: as.ledger.ID,
			Currency: "USD",
		})
		require.NoError(t, err)

		// Should succeed since no transactions exist
		err = accountBusiness.DeleteAccount(ctx, "delete-me-account")
		require.NoError(t, err)
	})
}

func (as *AccountBusinessSuite) TestDeleteAccountWithTransactions() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)
		as.setupFixtures(ctx, resources)

		accountBusiness := resources.AccountBusiness
		ledgerBusiness := resources.LedgerBusiness
		transactionBusiness := resources.TransactionBusiness

		// Create income ledger and accounts
		incomeLedger, err := ledgerBusiness.CreateLedger(ctx, &ledgerv1.CreateLedgerRequest{
			Id:   "del-income-ledger",
			Type: ledgerv1.LedgerType_INCOME,
		})
		require.NoError(t, err)

		_, err = accountBusiness.CreateAccount(ctx, &ledgerv1.CreateAccountRequest{
			Id:       "del-asset-acc",
			LedgerId: as.ledger.ID,
			Currency: "USD",
		})
		require.NoError(t, err)

		_, err = accountBusiness.CreateAccount(ctx, &ledgerv1.CreateAccountRequest{
			Id:       "del-income-acc",
			LedgerId: incomeLedger.GetId(),
			Currency: "USD",
		})
		require.NoError(t, err)

		// Create a transaction involving this account
		_, err = transactionBusiness.CreateTransaction(ctx, &ledgerv1.CreateTransactionRequest{
			Id:       "del-test-txn",
			Currency: "USD",
			Type:     ledgerv1.TransactionType_NORMAL,
			Entries: []*ledgerv1.TransactionEntry{
				{AccountId: "del-asset-acc", Credit: false, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
				{AccountId: "del-income-acc", Credit: true, Amount: &money.Money{CurrencyCode: "USD", Units: 100}},
			},
		})
		require.NoError(t, err)

		// Should fail since account has transactions
		err = accountBusiness.DeleteAccount(ctx, "del-asset-acc")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has transactions")
	})
}

func (as *AccountBusinessSuite) TestDeleteAccountEmptyID() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)

		err := resources.AccountBusiness.DeleteAccount(ctx, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "account ID is required")
	})
}

func (as *AccountBusinessSuite) TestEdgeCases() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)
		as.setupFixtures(ctx, resources)

		accountBusiness := resources.AccountBusiness

		// Test UpdateAccount with non-existent ID (triggers GetByID error path)
		_, err := accountBusiness.UpdateAccount(ctx, &ledgerv1.UpdateAccountRequest{
			Id: "non-existent-account-update",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"name": {Kind: &structpb.Value_StringValue{StringValue: "test"}},
				},
			},
		})
		require.Error(t, err)

		// Test UpdateAccount with empty ID
		_, err = accountBusiness.UpdateAccount(ctx, &ledgerv1.UpdateAccountRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "account ID is required")

		// Test GetAccount with empty ID
		_, err = accountBusiness.GetAccount(ctx, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "account ID is required")

		// Test SearchAccounts with invalid query (triggers search error result path)
		invalidQuery := &commonv1.SearchRequest{Query: "invalid json"}
		var foundAccounts []*ledgerv1.Account
		err = accountBusiness.SearchAccounts(
			ctx,
			invalidQuery,
			func(_ context.Context, batch []*ledgerv1.Account) error {
				foundAccounts = append(foundAccounts, batch...)
				return nil
			},
		)
		require.Error(t, err)

		// Test SearchAccounts with consumer error
		searchReq := &commonv1.SearchRequest{Query: "{}"}
		_, err = accountBusiness.CreateAccount(ctx, &ledgerv1.CreateAccountRequest{
			Id:       "consumer-err-acc",
			LedgerId: as.ledger.ID,
			Currency: "USD",
		})
		require.NoError(t, err)

		consumerErr := errors.New("consumer failed")
		err = accountBusiness.SearchAccounts(ctx, searchReq, func(_ context.Context, batch []*ledgerv1.Account) error {
			return consumerErr
		})
		require.Error(t, err)
		assert.Equal(t, consumerErr, err)

		// Test CreateAccount with empty ledger ID
		_, err = accountBusiness.CreateAccount(ctx, &ledgerv1.CreateAccountRequest{
			Id:       "no-ledger-acc",
			Currency: "USD",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ledger ID is required")

		// Test SearchAccounts with nil query (empty string triggers default)
		var emptyQueryAccounts []*ledgerv1.Account
		err = accountBusiness.SearchAccounts(
			ctx,
			&commonv1.SearchRequest{},
			func(_ context.Context, batch []*ledgerv1.Account) error {
				emptyQueryAccounts = append(emptyQueryAccounts, batch...)
				return nil
			},
		)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(emptyQueryAccounts), 1)
	})
}

func (as *AccountBusinessSuite) TestUpdateAccount() {
	as.WithTestDependencies(as.T(), func(t *testing.T, dep *definition.DependencyOption) {
		ctx, _, resources := as.CreateService(t, dep)
		as.setupFixtures(ctx, resources)

		accountBusiness := resources.AccountBusiness

		// Create an account first
		createAccountReq := &ledgerv1.CreateAccountRequest{
			Id:       "update-test-account",
			LedgerId: as.ledger.ID,
			Currency: "GBP",
		}

		_, err := accountBusiness.CreateAccount(ctx, createAccountReq)
		require.NoError(t, err, "Error creating account")

		// Update the account data
		updateAccountReq := &ledgerv1.UpdateAccountRequest{
			Id: "update-test-account",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"description": {Kind: &structpb.Value_StringValue{StringValue: "Updated description"}},
					"category":    {Kind: &structpb.Value_StringValue{StringValue: "Test category"}},
				},
			},
		}

		updatedAccount, err := accountBusiness.UpdateAccount(ctx, updateAccountReq)
		require.NoError(t, err, "Error updating account")
		require.NotNil(t, updatedAccount, "Updated account should not be nil")

		// Verify the update
		assert.Equal(t, "Updated description", updatedAccount.GetData().GetFields()["description"].GetStringValue())
		assert.Equal(t, "Test category", updatedAccount.GetData().GetFields()["category"].GetStringValue())
	})
}
