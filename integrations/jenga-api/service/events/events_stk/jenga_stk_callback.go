package events_stk

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
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
	logger := event.Service.L(ctx)

	if event.PaymentClient == nil {
		return errors.New("payment client not initialized")
	}

	callback := payload.(*models.StkCallback)

	logger.WithFields(map[string]interface{}{
		"transaction_ref": callback.Transaction,
		"telco_ref":       callback.Telco,
		"mobile_number":   callback.MobileNumber,
	}).Info("Processing STK callback")

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
			ContactId: callback.MobileNumber,
			Extras: map[string]string{
				"mobile_number": callback.MobileNumber,
			},
		},
		Recipient: &commonv1.ContactLink{
			ContactId: callback.Telco,
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
	receiveResponse, err := event.PaymentClient.Client.Receive(ctx, receiveRequest)
	if err != nil {
		logger.WithError(err).Error("failed to receive payment")
		return nil
	}

	// Log the receive response
	logger.WithField("receive_response", receiveResponse).Info("Received receive response from payment service")

	// Verify payment ID exists before proceeding
	if receiveResponse.Data == nil || receiveResponse.Data.GetId() == "" {
		logger.Error("received empty payment ID from payment service")
		return nil
	}

	// Add a small delay to give time for the payment to be saved to the database
	// This prevents the race condition where we try to update a payment that hasn't been saved yet
	time.Sleep(500 * time.Millisecond)
	
	// Determine payment status based on callback.Status
	paymentStatus := commonv1.STATUS_SUCCESSFUL
	if !callback.Status {
		paymentStatus = commonv1.STATUS_FAILED
		logger.Info("Callback indicates payment failure, setting status to FAILED")
	}
	
	//  status update use commonv1 StatusUpdateRequest
	statusUpdateRequest := &commonv1.StatusUpdateRequest{
		Id:     receiveResponse.Data.GetId(),
		State:  commonv1.STATE_ACTIVE,
		Status: paymentStatus,
		Extras: map[string]string{
			"raw_callback": string(callbackJSON),
		},
	}

	// Invoke the GRPC status update method
	statusUpdateResponse, err := event.PaymentClient.StatusUpdate(ctx, statusUpdateRequest)
	if err != nil {
		logger.WithError(err).Error("failed to update payment status")
		
		// If the first attempt fails due to timing, try again with a longer delay
		if strings.Contains(err.Error(), "no entity found") {
			logger.Info("First status update attempt failed, retrying after a delay")
			time.Sleep(1 * time.Second)
			
			statusUpdateResponse, err = event.PaymentClient.StatusUpdate(ctx, statusUpdateRequest)
			if err != nil {
				logger.WithError(err).Error("failed to update payment status after retry")
				return nil
			}
		} else {
			return nil
		}
	}
	// Log the status update response
	logger.WithField("status_update_response", statusUpdateResponse).Info("Status update response from payment service")

	return nil
}
