package events

import (
	"context"
	"errors"

	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments/service/models"
	"gorm.io/gorm/clause"

	"github.com/pitabwire/frame"
)

type PaymentLinkSave struct {
	Service    *frame.Service
	ProfileCli *profileV1.ProfileClient
}

func (e *PaymentLinkSave) Name() string {
	return "payment_link.save"
}

func (e *PaymentLinkSave) PayloadType() any {
	return &models.PaymentLink{}
}

func (e *PaymentLinkSave) Validate(ctx context.Context, payload any) error {
	logger := e.Service.Log(ctx).WithField("function", "PaymentLinkSave.Validate")

	paymentLink, ok := payload.(*models.PaymentLink)
	if !ok {
		logger.Error("Payload is not of type models.PaymentLink")
		return errors.New("payload is not of type models.PaymentLink")
	}

	// Log detailed ID information
	logger.
		WithField("paymentLink.ID", paymentLink.ID).
		WithField("paymentLink.GetID()", paymentLink.GetID()).
		WithField("paymentLink.BaseModel.ID", paymentLink.BaseModel.ID).
		Debug("Validating payment link ID")

	// Fix ID issues if possible
	if paymentLink.GetID() == "" {
		// If BaseModel ID is empty but explicit ID is set, try to use that
		if paymentLink.ID != "" {
			logger.Info("Using explicit ID field for validation")
			return nil
		}

		logger.Error("PaymentLink ID is not set and no fallback ID is available")
		return errors.New("payment link Id should already have been set")
	}

	// If we got here, the ID is valid
	logger.Debug("PaymentLink ID validation successful")
	return nil
}

func (e *PaymentLinkSave) Execute(ctx context.Context, payload any) error {
	paymentLink, ok := payload.(*models.PaymentLink)
	if !ok {
		return errors.New("payload is not of type models.PaymentLink")
	}

	logger := e.Service.Log(ctx).WithField("payload", paymentLink).WithField("type", e.Name())
	logger.Debug("handling event")

	// Attempt to save to database
	result := e.Service.DB(ctx, false).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(paymentLink)

	err := result.Error
	if err != nil {
		logger.WithError(err).Error("could not save payment link to db")
		// Return the error so the caller knows the save failed
		return err
	}

	logger.WithField("rows affected", result.RowsAffected).Debug("successfully saved record to db")
	return nil
}
