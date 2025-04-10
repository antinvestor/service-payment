package events

import (
	"context"
	"fmt"

	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/pitabwire/frame"
)

// JengaSTKUSSD handles STK/USSD push requests
type JengaSTKUSSD struct {
	Service       *frame.Service
	Client        *coreapi.Client
	PaymentClient *paymentV1.PaymentClient
}

// Name returns the name of the event handler
func (event *JengaSTKUSSD) Name() string {
	return "jenga_stk_ussd"
}

// PayloadType returns the type of payload this event expects
func (event *JengaSTKUSSD) PayloadType() any {
	return &models.STKUSSDRequest{}
}

// Validate validates the payload
func (event *JengaSTKUSSD) Validate(ctx context.Context, payload any) error {
	request, ok := payload.(*models.STKUSSDRequest)
	if !ok {
		return fmt.Errorf("invalid payload type")
	}

	// Basic validation
	if request.Merchant.AccountNumber == "" {
		return fmt.Errorf("merchant account number is required")
	}
	if request.Payment.Amount == "" {
		return fmt.Errorf("payment amount is required")
	}
	if request.Payment.MobileNumber == "" {
		return fmt.Errorf("mobile number is required")
	}

	return nil
}

// Execute handles the STK/USSD push request
func (event *JengaSTKUSSD) Execute(ctx context.Context, payload any) error {
	request := payload.(*models.STKUSSDRequest)
	
	logger := event.Service.L(ctx)

	// Generate bearer token for authorization
	token, err := event.Client.GenerateBearerToken()
	if err != nil {
		logger.WithError(err).Error("failed to generate bearer token")
		return fmt.Errorf("failed to generate bearer token: %v", err)
	}

	// Initiate STK/USSD push
	response, err := event.Client.InitiateSTKUSSD(*request, token.AccessToken)
	if err != nil {
		logger.WithError(err).Error("failed to initiate STK/USSD push")
		return fmt.Errorf("failed to initiate STK/USSD push: %v", err)
	}

	logger.WithField("response", response).Info("Successfully initiated STK/USSD push")

	// Create STK service payment event
	stkServicePayment := &JengaSTKServicePayment{
		Service:       event.Service,
		PaymentClient: event.PaymentClient,
	}

	// Execute STK service payment
	if err := stkServicePayment.Execute(ctx, request); err != nil {
		logger.WithError(err).Error("failed to execute STK service payment")
		return fmt.Errorf("failed to execute STK service payment: %v", err)
	}

	return nil
}
