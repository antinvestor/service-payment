//nolint:revive // package name matches directory structure
package eventscallback

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"buf.build/gen/go/antinvestor/payment/connectrpc/go/payment/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/payment/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/util"
	"github.com/pitabwire/util/decimalx"
	utilmoney "github.com/pitabwire/util/money"
)

type JengaStkCallback struct {
	PaymentClient paymentv1connect.PaymentServiceClient
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

	callbackJSON, err := json.Marshal(callback)
	if err != nil {
		logger.WithError(err).Error("failed to marshal callback")
		return nil
	}

	cbJSON := data.JSONMap{
		"additional_info": string(callbackJSON),
	}

	logger.WithField("callback", callback).Info("Received Jenga STK callback")

	amtDec, _ := decimalx.NewFromString(fmt.Sprintf("%g", callback.RequestAmount))
	amount := utilmoney.ToMoney(callback.Currency, amtDec)
	costDec, _ := decimalx.NewFromString(fmt.Sprintf("%g", callback.Charge))
	cost := utilmoney.ToMoney(callback.Currency, costDec)

	payment := &paymentv1.Payment{
		TransactionId: callback.Transaction,
		Amount:        amount,
		Cost:          cost,
		Extra:         cbJSON.ToProtoStruct(),
	}

	_, err = event.PaymentClient.Receive(ctx, connect.NewRequest(&paymentv1.ReceiveRequest{
		Data: payment,
	}))
	if err != nil {
		logger.WithError(err).Error("failed to process STK callback")
		return err
	}

	return nil
}
