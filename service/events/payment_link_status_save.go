package events

import (
	"context"
	"errors"

	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame"
	"gorm.io/gorm/clause"
)

type PaymentLinkStatusSave struct {
	Service    *frame.Service
	ProfileCli *profileV1.ProfileClient
}

func (e *PaymentLinkStatusSave) Name() string {
	return "paymentLinkStatus.save"
}

func (e *PaymentLinkStatusSave) PayloadType() any {
	return &models.PaymentLinkStatus{}
}

func (e *PaymentLinkStatusSave) Validate(ctx context.Context, payload any) error {
	logger := e.Service.L(ctx).WithField("function", "PaymentLinkStatusSave.Validate")

	paymentLinkStatus, ok := payload.(*models.PaymentLinkStatus)
	if !ok {
		logger.Error("Payload is not of type models.PaymentLinkStatus")
		return errors.New("payload is not of type models.PaymentLinkStatus")
	}

	// Log detailed ID information
	logger.WithFields(map[string]interface{}{
		"paymentLinkStatus.ID":      paymentLinkStatus.ID,
		"paymentLinkStatus.GetID()": paymentLinkStatus.GetID(),
		"paymentLinkStatus.PaymentLinkID": paymentLinkStatus.PaymentLinkID,
	}).Debug("Validating payment link status ID")

	// Check for ID validity
	if paymentLinkStatus.GetID() == "" {
		// If BaseModel ID is empty but explicit ID is set, try to use that
		if paymentLinkStatus.ID != "" {
			logger.Info("Using explicit ID field as fallback")
			return nil
		}

		// If PaymentLinkID is set but ID isn't, use PaymentLinkID as ID
		if paymentLinkStatus.PaymentLinkID != "" {
			paymentLinkStatus.ID = paymentLinkStatus.PaymentLinkID
			logger.Info("Setting ID from PaymentLinkID field")
			return nil
		}

		logger.Error("PaymentLinkStatus ID is not set and no fallback ID is available")
		return errors.New("paymentLinkStatus Id should already have been set")
	}

	// If we got here, the ID is valid
	logger.Debug("PaymentLinkStatus ID validation successful")
	return nil
}

func (e *PaymentLinkStatusSave) Execute(ctx context.Context, payload any) error {
	paymentLinkStatus := payload.(*models.PaymentLinkStatus)

	logger := e.Service.L(ctx).WithField("payload", paymentLinkStatus).WithField("type", e.Name())
	logger.Debug("handling event")

	result := e.Service.DB(ctx, false).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(paymentLinkStatus)

	err := result.Error
	if err != nil {
		logger.WithError(err).Warn("could not save payment link status to db")
		return err
	}
	logger.WithField("rows affected", result.RowsAffected).Debug("successfully saved record to db")

	return nil
}
