//nolint:dupl // Similar pattern for different models is acceptable
package events

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
	"github.com/pitabwire/util"
)

type PaymentLinkSave struct {
	paymentLinkRepo repository.PaymentLinkRepository
}

// NewPaymentLinkSave creates a new PaymentLinkSave event handler with the required dependencies.
func NewPaymentLinkSave(paymentLinkRepo repository.PaymentLinkRepository) *PaymentLinkSave {
	return &PaymentLinkSave{
		paymentLinkRepo: paymentLinkRepo,
	}
}

func (e *PaymentLinkSave) Name() string {
	return EventNamePaymentLinkSave
}

func (e *PaymentLinkSave) PayloadType() any {
	return &models.PaymentLink{}
}

func (e *PaymentLinkSave) Validate(ctx context.Context, payload any) error {
	logger := util.Log(ctx).WithField("function", "PaymentLinkSave.Validate")

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

	logger := util.Log(ctx).WithField("payload", paymentLink).WithField("type", e.Name())
	logger.Debug("handling event")

	// Attempt to save to database
	err := e.paymentLinkRepo.Create(ctx, paymentLink)
	if err != nil {
		logger.WithError(err).Error("could not save payment link to db")
		// Return the error so the caller knows the save failed
		return err
	}

	logger.Debug("successfully saved record to db")
	return nil
}
