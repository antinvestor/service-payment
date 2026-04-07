//nolint:revive // package name matches directory structure
package eventstillspay

import (
	"context"
	"errors"
	"fmt"

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

func (event *JengaTillsPay) Validate(_ context.Context, payload any) error {
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

	logger := util.Log(ctx).WithFields(map[string]any{
		"event":       event.Name(),
		"payment_ref": request.Payment.Ref,
		"till":        request.Merchant.Till,
		"amount":      request.Payment.Amount,
		"currency":    request.Payment.Currency,
		"partner_id":  request.Partner.ID,
	})
	logger.Info("processing tills pay")

	token, err := event.client.GenerateBearerToken(ctx)
	if err != nil {
		logger.WithError(err).Error("failed to generate bearer token for tills pay")
		return fmt.Errorf("generate bearer token: %w", err)
	}

	response, err := event.client.InitiateTillsPay(ctx, *request, token.AccessToken)
	if err != nil {
		logger.WithError(err).Error("tills pay initiation failed")
		return fmt.Errorf("initiate tills pay: %w", err)
	}

	logger.WithFields(map[string]any{
		"response_status": response.Status,
		"transaction_id":  response.TransactionID,
		"merchant_name":   response.MerchantName,
	}).Info("tills pay completed successfully")

	return nil
}
