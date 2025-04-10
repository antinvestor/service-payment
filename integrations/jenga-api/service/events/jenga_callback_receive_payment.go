
package events

import (
	"context"
	"encoding/json"
	"errors"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/service/models"
	"google.golang.org/genproto/googleapis/type/money"
	"github.com/pitabwire/frame"
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
	request := payload.(*models.CallbackRequest)

	if request.Transaction.Reference == "" {
		return errors.New("transaction reference is required")
	}

	return nil
}

func (event *JengaCallbackReceivePayment) Execute(ctx context.Context, payload any) error {

	if event.PaymentClient == nil {
		return errors.New("payment client not initialized")
	}

	callback := payload.(*models.CallbackRequest)

	// Extract relevant information from callback
	payment := &paymentV1.Payment{
		ReferenceId: callback.Transaction.Reference,
		Amount: &money.Money{
			Units: int64(callback.Transaction.Amount * 100), // convert to cents
			CurrencyCode: callback.Transaction.Currency,
		},
		Cost: &money.Money{
			Units: int64(callback.Transaction.ServiceCharge * 100), // convert to cents
			CurrencyCode: callback.Transaction.OrderCurrency,
		},
		Source: &commonv1.ContactLink{
			ContactId: callback.Customer.Reference,
			Extras: map[string]string{
				"mobile_number": callback.Customer.MobileNumber,
			},
			ProfileName: callback.Customer.Name,
		},
	    Recipient: &commonv1.ContactLink{
			ContactId: callback.Bank.Reference,
			Extras: map[string]string{
				"account": *callback.Bank.Account,
			},
		},
			

	}

	// Add any additional information from callback to extras
	extras := make(map[string]string)
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
	_, err = event.PaymentClient.Client.Receive(ctx, receiveRequest)
	return err
}
