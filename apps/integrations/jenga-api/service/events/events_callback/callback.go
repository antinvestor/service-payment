package events_callback

import (
	"context"
	"encoding/json"
	"errors"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/antinvestor/jenga-api/service/utility"
	"github.com/pitabwire/frame"
	"github.com/shopspring/decimal"
)

type JengaCallbackReceivePayment struct {
	Service       *frame.Service
	PaymentClient *paymentV1.PaymentClient
}

func (event *JengaCallbackReceivePayment) Name() string {
	return "jenga.callback.receive.payment"
}

func (event *JengaCallbackReceivePayment) PayloadType() any {
	return &models.CallbackRequest{}
}

func (event *JengaCallbackReceivePayment) Validate(ctx context.Context, payload any) error {
	req := payload.(*models.CallbackRequest)

	if req.Transaction.Reference == "" {
		return errors.New("transaction reference is required")
	}

	return nil
}

func (event *JengaCallbackReceivePayment) Execute(ctx context.Context, payload any) error {
	// Get logger first to avoid redefinition
	logger := event.Service.Log(ctx)

	if event.PaymentClient == nil {
		return errors.New("payment client not initialized")
	}

	req := payload.(*models.CallbackRequest)

	logger.WithField("callback", req).Info("Received Jenga callback for payment processing")

	// Create base payment structure
	amount := utility.ToMoney(req.Transaction.Currency, decimal.NewFromFloat(req.Transaction.Amount))
	cost := utility.ToMoney(req.Transaction.Currency, decimal.NewFromFloat(req.Transaction.ServiceCharge))
	payment := &paymentV1.Payment{
		Source: &commonv1.ContactLink{
			Detail: req.Customer.MobileNumber,
		},
		Recipient: &commonv1.ContactLink{
			Detail: req.Bank.Account,
		},
		TransactionId: req.Transaction.Reference,
		Amount:        &amount,
		Cost:          &cost,
	}

	var callbackJSON []byte
	var err error
	if callbackJSON, err = json.Marshal(req); err == nil {
		payment.Extra["additional_info"] = string(callbackJSON)
	}
	receiveRequest := &paymentV1.ReceiveRequest{
		Data: payment,
	}

	// Invoke the GRPC receive method
	_, err = event.PaymentClient.Client.Receive(ctx, receiveRequest)
	if err != nil {
		logger.WithError(err).Error("failed to receive payment")
		return err
	}
	return nil

}
