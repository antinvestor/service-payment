//nolint:revive // package name matches directory structure
package events_callback //nolint:staticcheck // underscore package name required by project structure

import (
	"context"
	"encoding/json"
	"errors"

	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/payment/v1"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/antinvestor/jenga-api/service/utility"
	"github.com/pitabwire/frame"
	"github.com/shopspring/decimal"
)

type JengaStkCallback struct {
	Service       *frame.Service
	PaymentClient *paymentv1.PaymentClient
}

func (event *JengaStkCallback) Name() string {
	return "jenga.callback.stk.payment"
}

func (event *JengaStkCallback) PayloadType() any {
	return &models.StkCallback{}
}

func (event *JengaStkCallback) Validate(ctx context.Context, payload any) error {
	callback, ok := payload.(*models.StkCallback)
	if !ok {
		return errors.New("invalid payload type")
	}

	if callback.Transaction == "" {
		return errors.New("transaction reference is required")
	}
	if callback.MobileNumber == "" {
		return errors.New("mobile number is required")
	}

	return nil
}

func (event *JengaStkCallback) Execute(ctx context.Context, payload any) error {
	logger := util.Log(ctx)

	if event.PaymentClient == nil {
		return errors.New("payment client not initialized")
	}

	callback, ok := payload.(*models.StkCallback)
	if !ok {
		return errors.New("invalid payload type")
	}
	logger.WithField("callback", callback).Info("Received Jenga STK callback")

	amount := utility.ToMoney(callback.Currency, decimal.NewFromFloat(callback.RequestAmount))
	cost := utility.ToMoney(callback.Currency, decimal.NewFromFloat(callback.Charge))

	payment := &paymentv1.Payment{
		TransactionId: callback.Transaction,
		Amount:        &amount,
		Cost:          &cost,
	}

	if callbackJSON, err := json.Marshal(callback); err == nil {
		payment.Extra["additional_info"] = string(callbackJSON)
	}
	_, err := event.PaymentClient.Client.Receive(ctx, &paymentv1.ReceiveRequest{
		Data: payment,
	})
	if err != nil {
		logger.WithError(err).Error("failed to process STK callback")
		return err
	}

	return nil
}
