package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/settings/v1/settingsv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/config"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/coreapi"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	frameEvents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/util"
)

const (
	defaultCurrency = "KES"
	defaultTelco    = "Safaricom"
	defaultPushType = "STK"
	dateFormat      = "2006-01-02"
	amountFormat    = "%.2f"
)

type promptHandler struct {
	credentialResolver
	statusEmitter
	jengaCli coreapi.JengaApiClient
}

// NewPromptHandler creates a queue worker for handling STK/USSD prompt requests.
func NewPromptHandler(
	eventsMan frameEvents.Manager,
	jengaCli coreapi.JengaApiClient,
	settingsCli settingsv1connect.SettingsServiceClient,
	cfg *config.JengaConfig,
) queue.SubscribeWorker {
	return &promptHandler{
		credentialResolver: credentialResolver{settingsCli: settingsCli, cfg: cfg},
		statusEmitter:      statusEmitter{eventsMan: eventsMan},
		jengaCli:           jengaCli,
	}
}

func (h *promptHandler) Handle(ctx context.Context, headers map[string]string, payload []byte) error {
	logger := util.Log(ctx).WithField("handler", "jenga.prompt")
	defer logger.Release()

	// The payment service publishes models.Prompt as JSON (Go struct, not protobuf)
	var prompt models.Prompt
	if err := json.Unmarshal(payload, &prompt); err != nil {
		logger.WithError(err).Error("failed to unmarshal prompt")
		return nil // non-retriable
	}

	promptID := prompt.ID
	if promptID == "" {
		promptID = prompt.GetID()
	}

	logger = logger.WithFields(map[string]any{
		"prompt_id": promptID,
		"mobile":    prompt.SourceContactID,
	})
	logger.Info("processing STK prompt request")

	creds, err := h.extractCredentials(ctx, headers)
	if err != nil {
		logger.WithError(err).Error("failed to resolve Jenga credentials")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       err.Error(),
			"entity_type": "prompt",
		})
		return nil
	}

	// Extract transaction ref and other extras
	transactionRef, _ := prompt.Extra["transaction_ref"].(string)
	if transactionRef == "" {
		logger.Error("transaction reference is missing")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       "transaction reference is required",
			"entity_type": "prompt",
		})
		return nil
	}

	logger = logger.WithField("transaction_ref", transactionRef)

	currency := getStringWithDefault(prompt.Extra, "currency", defaultCurrency)
	telco := getStringWithDefault(prompt.Extra, "telco", defaultTelco)
	pushType := getStringWithDefault(prompt.Extra, "pushType", defaultPushType)

	var amountStr string
	if prompt.Amount != nil {
		amtFloat, _ := strconv.ParseFloat(prompt.Amount.String(), 64)
		amountStr = fmt.Sprintf(amountFormat, amtFloat)
	}

	// Parse account info from the prompt
	account := prompt.Account
	accountNumber, _ := account["accountNumber"].(string)
	countryCode, _ := account["countryCode"].(string)
	accountName, _ := account["name"].(string)

	stkRequest := models.STKUSSDRequest{
		Merchant: models.Merchant{
			AccountNumber: accountNumber,
			CountryCode:   countryCode,
			Name:          accountName,
		},
		Payment: models.Payment{
			Ref:          transactionRef,
			Amount:       amountStr,
			Currency:     currency,
			Telco:        telco,
			MobileNumber: prompt.SourceContactID,
			Date:         time.Now().Format(dateFormat),
			CallBackUrl:  creds.CallbackURL,
			PushType:     pushType,
		},
		ID: promptID,
	}

	token, err := h.jengaCli.GenerateBearerToken(ctx)
	if err != nil {
		logger.WithError(err).Error("failed to generate bearer token for STK push")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":       fmt.Sprintf("generate bearer token: %v", err),
			"entity_type": "prompt",
		})
		return nil
	}

	resp, err := h.jengaCli.InitiateSTKUSSD(ctx, stkRequest, token.AccessToken)
	if err != nil {
		logger.WithError(err).Error("STK/USSD push initiation failed")
		h.emitStatus(ctx, promptID, "", commonv1.STATUS_FAILED, map[string]any{
			"error":           fmt.Sprintf("initiate STK push: %v", err),
			"transaction_ref": transactionRef,
			"entity_type":     "prompt",
		})
		return nil
	}

	logger.WithFields(map[string]any{
		"response_transaction_id": resp.TransactionID,
		"response_status":         resp.Status,
		"response_message":        resp.Message,
	}).Info("STK/USSD push initiated successfully")

	h.emitStatus(ctx, promptID, resp.TransactionID, commonv1.STATUS_IN_PROCESS, map[string]any{
		"transaction_ref": transactionRef,
		"transaction_id":  resp.TransactionID,
		"response_code":   resp.Code,
		"message":         resp.Message,
		"entity_type":     "prompt",
	})

	return nil
}

func getStringWithDefault(extras map[string]interface{}, key, defaultValue string) string {
	if extras == nil {
		return defaultValue
	}
	if val, ok := extras[key].(string); ok && val != "" {
		return val
	}
	return defaultValue
}
