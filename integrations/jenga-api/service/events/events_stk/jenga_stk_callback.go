package events_stk

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/pitabwire/frame"
	"google.golang.org/genproto/googleapis/type/money"
)

type JengaCallbackReceivePayment struct {
	Service       *frame.Service
	PaymentClient *paymentV1.PaymentClient
}

func (event *JengaCallbackReceivePayment) Name() string {
	return "jenga.callback.receive.payment"
}

func (event *JengaCallbackReceivePayment) PayloadType() any {
	return &models.StkCallback{}
}

func (event *JengaCallbackReceivePayment) Validate(ctx context.Context, payload any) error {
	callback := payload.(*models.StkCallback)

	if callback.Transaction == "" {
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

	callback := payload.(*models.StkCallback)
	//TODO put payload to an extra

	logger.WithField("callback", callback).Info("Received Jenga STK callback for payment processing")

	// Extract relevant information from callback
	payment := &paymentV1.Payment{
		// Keep the original transaction reference - we'll handle lookup/mapping in the payment service
		TransactionId: callback.Transaction,
		Amount: &money.Money{
			Units:        int64(callback.DebitedAmount * 100), // convert to cents
			CurrencyCode: callback.Currency,
		},
		Cost: &money.Money{
			Units:        int64(callback.Charge * 100), // convert to cents
			CurrencyCode: callback.Currency,
		},
		Source: &commonv1.ContactLink{
			Detail: callback.MobileNumber,
			Extras: map[string]string{
				"mobile_number": callback.MobileNumber,
				"telco": callback.Telco,
			},
		},
		Recipient: &commonv1.ContactLink{
		    Detail: callback.Telco,
			Extras: map[string]string{
				"telco": callback.Telco,
			},
		},
	}

	// Add any additional information from callback to extras
	extras := make(map[string]string)

	// Add key data to help with reference handling
	extras["telco_ref"] = callback.Telco
	extras["mobile_number"] = callback.MobileNumber
	extras["update_type"] = "payment" // Explicitly indicate this is a payment update
	extras["callback_time"] = time.Now().Format(time.RFC3339)

	// Marshal the full callback to JSON and store it in extras
	callbackJSON, err := json.Marshal(callback)
	if err == nil {
		extras["raw_callback"] = string(callbackJSON)
	}

	payment.Extra = extras
	receiveRequest := &paymentV1.ReceiveRequest{
		Data: payment,
	}

	// Invoke the GRPC receive method
	_ , err = event.PaymentClient.Client.Receive(ctx, receiveRequest)
	if err != nil {
		logger.WithError(err).Error("failed to receive payment")
		return err
	}
	return nil


}
