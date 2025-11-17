//nolint:revive // package name matches directory structure
package eventstillspay

import (
	"context"
	"errors"

	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/coreapi"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/pitabwire/util"
)

type JengaTillsPay struct {
	client coreapi.JengaApiClient
}

// NewJengaTillsPay creates a new tills pay handler with dependencies.
func NewJengaTillsPay(
	client coreapi.JengaApiClient,
) *JengaTillsPay {
	return &JengaTillsPay{
		client: client,
	}
}

func (event *JengaTillsPay) Name() string {
	return "jenga.tills.pay"
}

func (event *JengaTillsPay) PayloadType() any {
	return &models.TillsPayRequest{}
}

func (event *JengaTillsPay) Validate(ctx context.Context, payload any) error {
	request, ok := payload.(*models.TillsPayRequest)
	if !ok {
		return errors.New("invalid payload type")
	}

	if request.Merchant.Till == "" {
		return errors.New("merchant.till is required")
	}
	if request.Payment.Ref == "" {
		return errors.New("payment.ref is required")
	}
	if request.Payment.Amount == "" {
		return errors.New("payment.amount is required")
	}
	if request.Payment.Currency == "" {
		return errors.New("payment.currency is required")
	}
	if request.Partner.ID == "" {
		return errors.New("partner.id is required")
	}
	if request.Partner.Ref == "" {
		return errors.New("partner.ref is required")
	}

	return nil
}

func (event *JengaTillsPay) Execute(ctx context.Context, payload any) error {
	request, ok := payload.(*models.TillsPayRequest)
	if !ok {
		return errors.New("invalid payload type")
	}

	logger := util.Log(ctx).WithField("type", event.Name()).WithField("TillsPayRequest", request)
	logger.WithField("request", request).Debug("processing tills pay")

	// Generate bearer token for authorization
	token, err := event.client.GenerateBearerToken()
	if err != nil {
		logger.WithError(err).Error("failed to generate bearer token")
		return err
	}
	logger.WithField("token", token.AccessToken).Info("generated token for tills pay")

	// Initiate the tills/pay API call
	resp, err := event.client.InitiateTillsPay(*request, token.AccessToken)
	if err != nil {
		logger.WithError(err).Error("failed to initiate tills pay")
		return err
	}
	logger.WithField("response", resp).Info("tills pay response")

	return nil
}
