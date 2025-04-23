package events

import (
	"context"
	"fmt"
	"time"

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

	// Get logger first to avoid redefinition
	logger := event.Service.L(ctx)

	// Generate a unique 6-character reference
	// We'll use the current time to help make it unique
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	// Use the last 6 digits of the timestamp (or fewer if needed)
	// This gives a rolling unique ID that cycles every million milliseconds (16.6 minutes)
	timeComponent := timestamp % 1000000

	originalRef := request.Payment.Ref

	// Generate the new reference (format: A12345, where A is alphabetic and 12345 are numeric)
	// This creates references like A12345, B56789, etc.
	asciiChar := 65 + ((timestamp / 1000000) % 26) // 65 is ASCII 'A', rotating through 26 letters
	request.Payment.Ref = fmt.Sprintf("%c%05d", rune(asciiChar), timeComponent%100000)

	logger.WithFields(map[string]interface{}{
		"original_ref": originalRef,
		"new_ref":      request.Payment.Ref,
	}).Info("Generated unique 6-character transaction reference")

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
	
	logger.WithField("response", response).Info("STK/USSD push response received")

	return nil
}
