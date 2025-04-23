package events

import (
	"context"
	"encoding/json"
	"errors"

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

	// Get logger first to avoid redefinition
	logger := event.Service.L(ctx)

	if event.PaymentClient == nil {
		return errors.New("payment client not initialized")
	}

	callback := payload.(*models.StkCallback)

	// Extract relevant information from callback
	payment := &paymentV1.Payment{
		ReferenceId: callback.Transaction,
		Amount: &money.Money{
			Units:        int64(callback.DebitedAmount * 100), // convert to cents
			CurrencyCode: callback.Currency,
		},
		Cost: &money.Money{
			Units:        int64(callback.Charge * 100), // convert to cents
			CurrencyCode: callback.Currency,
		},
		Source: &commonv1.ContactLink{
			ContactId: callback.MobileNumber,
			Extras: map[string]string{
				"mobile_number": callback.MobileNumber,
			},
		},
		Recipient: &commonv1.ContactLink{
			ContactId: callback.Telco,
			Extras: map[string]string{
				"account": callback.Telco,
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
	receiveResponse, err := event.PaymentClient.Client.Receive(ctx, receiveRequest)
	if err != nil {
		return err
	}

	// Log the receive response
	logger.WithField("receive_response", receiveResponse).Info("Received receive response from payment service")



	//  status update use commonv1 StatusUpdateRequest
	
	statusUpdateRequest := &commonv1.StatusUpdateRequest{
		Id:     receiveResponse.Data.Id,
		State:  commonv1.STATE_ACTIVE,
		Status: commonv1.STATUS_SUCCESSFUL,
		Extras: map[string]string{
			"raw_callback": string(callbackJSON),
		},
	}

	// Invoke the GRPC status update method
	statusUpdateResponse, err := event.PaymentClient.Client.StatusUpdate(ctx, statusUpdateRequest)
	if err != nil {
		return err
	}

	// Log the status update response
	logger.WithField("status_update_response", statusUpdateResponse).Info("Status update response from payment service")

	return nil
}
