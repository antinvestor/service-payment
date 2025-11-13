package events

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/antinvestor/service-payments/service/repository"
	"github.com/pitabwire/util"
)

const EventNameAccountSave = "account.save"

type AccountSave struct {
	accountRepo repository.AccountRepository
}

func (e *AccountSave) Name() string {
	return EventNameAccountSave
}

func (e *AccountSave) PayloadType() any {
	return &models.Account{}
}

func (e *AccountSave) Validate(ctx context.Context, payload any) error {
	logger := util.Log(ctx).WithField("function", "AccountSave.Validate")

	account, ok := payload.(*models.Account)
	if !ok {
		logger.Error("Payload is not of type models.Account")
		return errors.New("payload is not of type models.Account")
	}

	if account.ID == "" {
		logger.Error("Account ID is not set")
		return errors.New("account ID should already have been set")
	}

	logger.Debug("Account ID validation successful")
	return nil
}

func (e *AccountSave) Execute(ctx context.Context, payload any) error {
	account, ok := payload.(*models.Account)
	if !ok {
		return errors.New("payload is not of type models.Account")
	}

	logger := util.Log(ctx).WithField("payload", account).WithField("type", e.Name())
	logger.Debug("handling event")

	if account.Version > 0 {
		return nil
	}

	err := e.accountRepo.Create(ctx, account)
	if err != nil {
		logger.WithError(err).Error("could not save account to db")
		return err
	}

	logger.Debug("successfully saved record to db")
	return nil
}
