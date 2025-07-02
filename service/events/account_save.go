package events

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame"
	"gorm.io/gorm/clause"
)

type AccountSave struct {
	Service *frame.Service
}

func (e *AccountSave) Name() string {
	return "account.save"
}

func (e *AccountSave) PayloadType() any {
	return &models.Account{}
}

func (e *AccountSave) Validate(ctx context.Context, payload any) error {
	logger := e.Service.Log(ctx).WithField("function", "AccountSave.Validate")

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
	account := payload.(*models.Account)

	logger := e.Service.Log(ctx).WithField("payload", account).WithField("type", e.Name())
	logger.Debug("handling event")

	result := e.Service.DB(ctx, false).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(account)

	err := result.Error
	if err != nil {
		logger.WithError(err).Error("could not save account to db")
		return err
	}

	logger.WithField("rows affected", result.RowsAffected).Debug("successfully saved record to db")
	return nil
}
