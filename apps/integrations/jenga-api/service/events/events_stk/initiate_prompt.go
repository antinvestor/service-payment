//nolint:revive // package name matches directory structure
package events_stk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/pitabwire/frame"
	"gorm.io/datatypes"
)

const (
	defaultCurrency  = "KES"
	defaultTelco     = "Safaricom"
	defaultPushType  = "STK"
	dateFormat       = "2006-01-02"
	amountFormat     = "%.2f"
	updateTypePrompt = "prompt"
	statusActive     = commonv1.STATE_ACTIVE
	statusFailed     = commonv1.STATUS_FAILED
	statusSuccessful = commonv1.STATUS_SUCCESSFUL
)

// InitiatePrompt handles the initiate.prompt events.
type InitiatePrompt struct {
	Service       *frame.Service
	Client        coreapi.JengaApiClient
	PaymentClient paymentV1.PaymentClient
	CallbackURL   string
}

// NewInitiatePrompt creates a new InitiatePrompt handler with dependencies.
func NewInitiatePrompt(service *frame.Service, client coreapi.JengaApiClient, paymentClient paymentV1.PaymentClient, callbackURL string) *InitiatePrompt {
	return &InitiatePrompt{
		Service:       service,
		Client:        client,
		PaymentClient: paymentClient,
		CallbackURL:   callbackURL,
	}
}

// Name returns the name of the event handler.
func (h *InitiatePrompt) Name() string {
	return "initiate.prompt"
}

// PayloadType returns the type of payload this event expects.
func (h *InitiatePrompt) PayloadType() any {
	return &models.Prompt{}
}

// Validate validates the payload.
func (h *InitiatePrompt) Validate(ctx context.Context, payload any) error {
	prompt, ok := payload.(*models.Prompt)
	if !ok {
		return errors.New("invalid payload type, expected *models.Prompt")
	}

	if prompt.ID == "" {
		return errors.New("prompt ID is required")
	}
	if !prompt.Amount.Valid {
		return errors.New("payment amount is required")
	}
	if prompt.SourceContactID == "" {
		return errors.New("source contact ID (mobile number) is required")
	}

	return nil
}

// Handle implements the frame.SubscribeWorker interface.
func (h *InitiatePrompt) Handle(ctx context.Context, metadata map[string]string, message []byte) error {
	payload := h.PayloadType()
	if err := json.Unmarshal(message, payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if err := h.Validate(ctx, payload); err != nil {
		return fmt.Errorf("payload validation failed: %w", err)
	}

	return h.Execute(ctx, payload)
}

// Execute handles the prompt and initiates the STK/USSD push request.
func (h *InitiatePrompt) Execute(ctx context.Context, payload any) error {
	prompt, ok := payload.(*models.Prompt)
	if !ok {
		return errors.New("invalid payload type, expected *models.Prompt")
	}

	logger := h.Service.Log(ctx).WithField("promptId", prompt.ID)
	logger.Info("Processing initiate.prompt event")

	account, err := parseAccountInfo(prompt.Account)
	if err != nil {
		logger.WithError(err).Error("failed to parse account info")
		return fmt.Errorf("parse account info: %w", err)
	}

	transactionRef, ok := getStringFromExtra(prompt.Extra, "transaction_ref")
	if !ok {
		logger.Error("transaction reference is missing or invalid")
		return errors.New("transaction reference is required")
	}

	currency := getStringWithDefault(prompt.Extra, "currency", defaultCurrency)
	telco := getStringWithDefault(prompt.Extra, "telco", defaultTelco)
	pushType := getStringWithDefault(prompt.Extra, "pushType", defaultPushType)

	amountStr := fmt.Sprintf(amountFormat, prompt.Amount.Decimal.InexactFloat64())
	currentDate := time.Now().Format(dateFormat)

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
			CallBackUrl:  h.CallbackURL,
			PushType:     pushType,
		},
		ID: prompt.ID,
	}

	logger.WithField("stkRequest", stkRequest).Info("Prepared STK request")

	token, err := h.Client.GenerateBearerToken()
	if err != nil {
		logger.WithError(err).Error("failed to generate bearer token")
		return h.handleError(ctx, prompt.ID, transactionRef,
			fmt.Errorf("generate bearer token: %w", err))
	}

	response, err := h.Client.InitiateSTKUSSD(*stkRequest, token.AccessToken)
	if err != nil {
		logger.WithError(err).Error("failed to initiate STK/USSD push")
		return h.handleError(ctx, prompt.ID, transactionRef,
			fmt.Errorf("initiate STK/USSD push: %w", err))
	}

	logger.WithField("response", response).Info("STK/USSD push response received")

	if err := h.updateStatus(ctx, prompt.ID, transactionRef, response.TransactionID, response.Message); err != nil {
		logger.WithError(err).Error("failed to update payment status")
		return fmt.Errorf("update payment status: %w", err)
	}

	return nil
}

// parseAccountInfo unmarshals the account JSON from the prompt.
func parseAccountInfo(accountJSON datatypes.JSON) (*models.Account, error) {
	var account models.Account
	if err := json.Unmarshal(accountJSON, &account); err != nil {
		return nil, fmt.Errorf("unmarshal account JSON: %w", err)
	}
	return &account, nil
}

// getStringFromExtra safely extracts a string value from the extras map.
func getStringFromExtra(extras map[string]interface{}, key string) (string, bool) {
	if extras == nil {
		return "", false
	}
	val, ok := extras[key].(string)
	return val, ok && val != ""
}

// getStringWithDefault extracts a string with a default value if not found.
func getStringWithDefault(extras map[string]interface{}, key, defaultValue string) string {
	if val, ok := getStringFromExtra(extras, key); ok {
		return val
	}
	return defaultValue
}

// handleError updates the status to failed and returns the error.
func (h *InitiatePrompt) handleError(ctx context.Context, promptID, transactionRef string, err error) error {
	logger := h.Service.Log(ctx).WithField("promptId", promptID)
	statusUpdateRequest := &commonv1.StatusUpdateRequest{
		Id:     promptID,
		State:  statusActive,
		Status: statusFailed,
		Extras: map[string]string{
			"update_type":     updateTypePrompt,
			"transaction_ref": transactionRef,
			"error":           err.Error(),
		},
	}

	if _, updateErr := h.PaymentClient.StatusUpdate(ctx, statusUpdateRequest); updateErr != nil {
		logger.WithError(updateErr).Error("failed to update payment status")
		return fmt.Errorf("update payment status: %w", updateErr)
	}

	return err
}

// updateStatus updates the payment status to successful.
func (h *InitiatePrompt) updateStatus(ctx context.Context, promptID, transactionRef, transactionID, message string) error {
	statusUpdateRequest := &commonv1.StatusUpdateRequest{
		Id:     promptID,
		State:  statusActive,
		Status: statusSuccessful,
		Extras: map[string]string{
			"update_type":     updateTypePrompt,
			"transaction_ref": transactionRef,
			"transaction_id":  transactionID,
			"message":         message,
		},
	}

	_, err := h.PaymentClient.StatusUpdate(ctx, statusUpdateRequest)
	if err != nil {
		return fmt.Errorf("payment client status update: %w", err)
	}

	return nil
}
