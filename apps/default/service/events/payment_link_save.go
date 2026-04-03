//nolint:dupl // Similar pattern for different models is acceptable
package events

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
	"github.com/pitabwire/frame/data"
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

func (e *PaymentLinkSave) Validate(_ context.Context, payload any) error {
	paymentLink, ok := payload.(*models.PaymentLink)
	if !ok {
		return errors.New("payload is not of type models.PaymentLink")
	}

	if paymentLink.GetID() == "" {
		if paymentLink.ID != "" {
			return nil
		}
		return errors.New("payment link Id should already have been set")
	}

	return nil
}

func (e *PaymentLinkSave) Execute(ctx context.Context, payload any) error {
	paymentLink, ok := payload.(*models.PaymentLink)
	if !ok {
		return errors.New("payload is not of type models.PaymentLink")
	}

	logger := util.Log(ctx).WithFields(map[string]any{"payment_link_id": paymentLink.ID, "type": e.Name()})
	logger.Debug("handling event")

	// Attempt to save to database
	err := e.paymentLinkRepo.Create(ctx, paymentLink)
	if err != nil {
		if data.ErrorIsDuplicateKey(err) {
			logger.Debug("record already exists, skipping duplicate")
			return nil
		}
		logger.WithError(err).Warn("could not save payment link to db")
		return err
	}

	logger.Debug("successfully saved record to db")
	return nil
}
