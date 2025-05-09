package events_stk

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/service/coreapi"
	models "github.com/antinvestor/jenga-api/service/models"
	"github.com/pitabwire/frame"
)

// InitiatePrompt handles the initiate.prompt events
type InitiatePrompt struct {
	Service       *frame.Service
	Client        coreapi.JengaApiClient
	PaymentClient *paymentV1.PaymentClient
}

// Name returns the name of the event handler
func (event *InitiatePrompt) Name() string {
	return "initiate.prompt"
}

// PayloadType returns the type of payload this event expects
func (event *InitiatePrompt) PayloadType() any {
	return &models.Prompt{}
}

// Validate validates the payload
func (event *InitiatePrompt) Validate(ctx context.Context, payload any) error {
	prompt, ok := payload.(*models.Prompt)
	if !ok {
		return fmt.Errorf("invalid payload type, expected Prompt")
	}

	// Basic validation
	if prompt.ID == "" {
		return fmt.Errorf("prompt ID is required")
	}
	if !prompt.Amount.Valid {
		return fmt.Errorf("payment amount is required")
	}
	if prompt.SourceContactID == "" {
		return fmt.Errorf("source contact ID (mobile number) is required")
	}

	return nil
}

// Handle implements the frame.SubscribeWorker interface
func (event *InitiatePrompt) Handle(ctx context.Context, metadata map[string]string, message []byte) error {
	// Create a new payload instance
	payload := event.PayloadType()
	prompt, ok := payload.(*models.Prompt)
	if !ok {
		return fmt.Errorf("invalid payload type, expected Prompt")
	}

	// Unmarshal the message into the payload
	if err := json.Unmarshal(message, prompt); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	// Validate the payload
	if err := event.Validate(ctx, prompt); err != nil {
		return fmt.Errorf("payload validation failed: %v", err)
	}

	// Execute the business logic
	return event.Execute(ctx, prompt)
}

// Execute handles the prompt and initiates the STK/USSD push request
func (event *InitiatePrompt) Execute(ctx context.Context, payload any) error {
	prompt, ok := payload.(*models.Prompt)
	if !ok {
		return fmt.Errorf("invalid payload type")
	}

	// Get logger
	logger := event.Service.L(ctx).WithField("promptId", prompt.ID)
	logger.Info("Processing initiate.prompt event")

	// Extract account information from JSON
	var account models.Account
	if err := json.Unmarshal(prompt.Account, &account); err != nil {
		logger.WithError(err).Error("failed to unmarshal account JSON")
		return fmt.Errorf("failed to parse account info: %v", err)
	}

	// Get transaction reference from Extra
	transactionRef, ok := prompt.Extra["transaction_ref"].(string)
	if !ok || transactionRef == "" {
		logger.Error("transaction reference is missing or invalid")
		return fmt.Errorf("transaction reference is required")
	}

	// Get currency from Extra
	currency, ok := prompt.Extra["currency"].(string)
	if !ok || currency == "" {
		currency = "KES" // Default to KES if not provided
		logger.WithField("currency", currency).Info("Using default currency")
	}

	// Get telco from Extra
	telco, ok := prompt.Extra["telco"].(string)
	if !ok || telco == "" {
		telco = "Safaricom" // Default to Safaricom if not provided
		logger.WithField("telco", telco).Info("Using default telco")
	}

	// Get pushType from Extra
	pushType, ok := prompt.Extra["pushType"].(string)
	if !ok || pushType == "" {
		pushType = "STK" // Default to STK if not provided
		logger.WithField("pushType", pushType).Info("Using default push type")
	}

	// Format the amount for the API
	amountStr := fmt.Sprintf("%.2f", prompt.Amount.Decimal.InexactFloat64())

	// Get callback URL from environment variable
	callbackURL := os.Getenv("CALLBACK_URL")
	if callbackURL == "" {
		callbackURL = "https://callback.example.com" // Default callback URL
		logger.WithField("callbackURL", callbackURL).Warn("Using default callback URL")
	}

	// Format current date for the API
	currentDate := time.Now().Format("2006-01-02")

	// Prepare the STK/USSD push request
	stkRequest := &models.STKUSSDRequest{
		Merchant: models.Merchant{
			AccountNumber: account.AccountNumber,
			CountryCode:   account.CountryCode,
			Name:          account.Name,
		},
		Payment: models.Payment{
			Ref:          transactionRef,
			Amount:       amountStr,
			Currency:     currency,
			Telco:        telco,
			MobileNumber: prompt.SourceContactID,
			Date:         currentDate,
			CallBackUrl:  callbackURL,
			PushType:     pushType,
		},
		ID: prompt.ID,
	}

	logger.WithField("stkRequest", stkRequest).Info("Prepared STK request")

	// Generate bearer token for authorization
	token, err := event.Client.GenerateBearerToken()
	if err != nil {
		logger.WithError(err).Error("failed to generate bearer token")
		// Update status to failed
		statusUpdateRequest := &commonv1.StatusUpdateRequest{
			Id:     prompt.ID,
			State:  commonv1.STATE_ACTIVE,
			Status: commonv1.STATUS_FAILED,
			Extras: map[string]string{
				"update_type":     "prompt",
				"transaction_ref": transactionRef,
				"error":           fmt.Sprintf("failed to generate token: %v", err),
			},
		}
		_, updateErr := event.PaymentClient.StatusUpdate(ctx, statusUpdateRequest)
		if updateErr != nil {
			logger.WithError(updateErr).Error("failed to update payment status")
		}
		return fmt.Errorf("failed to generate bearer token: %v", err)
	}

	// Initiate STK/USSD push
	response, err := event.Client.InitiateSTKUSSD(*stkRequest, token.AccessToken)
	if err != nil {
		logger.WithError(err).Error("failed to initiate STK/USSD push")
		// Update status to failed
		statusUpdateRequest := &commonv1.StatusUpdateRequest{
			Id:     prompt.ID,
			State:  commonv1.STATE_ACTIVE,
			Status: commonv1.STATUS_FAILED,
			Extras: map[string]string{
				"update_type":     "prompt",
				"transaction_ref": transactionRef,
				"error":           err.Error(),
			},
		}
		_, updateErr := event.PaymentClient.StatusUpdate(ctx, statusUpdateRequest)
		if updateErr != nil {
			logger.WithError(updateErr).Error("failed to update payment status")
		}
		return fmt.Errorf("failed to initiate STK/USSD push: %v", err)
	}

	logger.WithField("response", response).Info("STK/USSD push response received")

	// Update status to processing
	statusUpdateRequest := &commonv1.StatusUpdateRequest{
		Id:     prompt.ID,
		State:  commonv1.STATE_ACTIVE,
		Status: commonv1.STATUS_SUCCESSFUL,
		Extras: map[string]string{
			"update_type":     "prompt",
			"transaction_ref": transactionRef,
			"transaction_id":  response.TransactionID,
			"message":         response.Message,
		},
	}
	_, err = event.PaymentClient.StatusUpdate(ctx, statusUpdateRequest)
	if err != nil {
		logger.WithError(err).Error("failed to update payment status")
	}

	return nil
}

