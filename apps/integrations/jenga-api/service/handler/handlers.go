package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1connect "buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/util"
	"github.com/pitabwire/util/decimalx"
	utilmoney "github.com/pitabwire/util/money"
)

// WebhookServer handles incoming callbacks from the Jenga API.
type WebhookServer struct {
	paymentCli paymentv1connect.PaymentServiceClient
}

// NewWebhookServer creates a new webhook server for Jenga API callbacks.
func NewWebhookServer(
	paymentCli paymentv1connect.PaymentServiceClient,
) *WebhookServer {
	return &WebhookServer{
		paymentCli: paymentCli,
	}
}

// NewRouter creates the HTTP router for Jenga webhook endpoints.
func (ws *WebhookServer) NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /receivepayments", ws.HandleStkCallback)
	mux.HandleFunc("POST /webhook/jenga/callback", ws.HandleGeneralCallback)
	return mux
}

// HandleStkCallback processes STK push callback notifications from Jenga.
func (ws *WebhookServer) HandleStkCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := util.Log(ctx).WithField("handler", "HandleStkCallback")

	var callback models.StkCallback
	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		logger.WithError(err).Error("failed to decode STK callback")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if callback.Transaction == "" || callback.MobileNumber == "" || callback.Currency == "" {
		logger.Error("missing required fields in STK callback")
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	logger = logger.WithFields(map[string]any{
		"transaction_ref": callback.Transaction,
		"mobile_number":   callback.MobileNumber,
		"amount":          callback.RequestAmount,
		"currency":        callback.Currency,
		"stk_status":      callback.Status,
	})
	logger.Info("received STK callback from Jenga")

	callbackJSON, _ := json.Marshal(callback)

	// Determine final status from callback
	status := commonv1.STATUS_SUCCESSFUL
	if !callback.Status {
		status = commonv1.STATUS_FAILED
	}

	amtDec, _ := decimalx.NewFromString(fmt.Sprintf("%g", callback.RequestAmount))
	amount := utilmoney.ToMoney(callback.Currency, amtDec)
	costDec, _ := decimalx.NewFromString(fmt.Sprintf("%g", callback.Charge))
	cost := utilmoney.ToMoney(callback.Currency, costDec)

	cbJSON := data.JSONMap{"additional_info": string(callbackJSON)}

	payment := &paymentv1.Payment{
		TransactionId: callback.Transaction,
		Amount:        amount,
		Cost:          cost,
		Extra:         cbJSON.ToProtoStruct(),
	}

	// Report the payment back to the payment service
	_, err := ws.paymentCli.Receive(ctx, connect.NewRequest(&paymentv1.ReceiveRequest{
		Data: payment,
	}))
	if err != nil {
		logger.WithError(err).Error("failed to forward STK callback to payment service")
		http.Error(w, "Failed to process callback", http.StatusInternalServerError)
		return
	}

	// Also update the prompt status
	_, err = ws.paymentCli.StatusUpdate(ctx, connect.NewRequest(&commonv1.StatusUpdateRequest{
		Id:         callback.Transaction,
		State:      commonv1.STATE_ACTIVE,
		Status:     status,
		ExternalId: callback.Transaction,
		Extras:     cbJSON.ToProtoStruct(),
	}))
	if err != nil {
		logger.WithError(err).Warn("failed to update prompt status after STK callback")
	}

	logger.Info("STK callback processed successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck // best-effort response
		"status":  "success",
		"message": "Callback received successfully",
	})
}

// HandleGeneralCallback processes general payment callback notifications from Jenga.
func (ws *WebhookServer) HandleGeneralCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := util.Log(ctx).WithField("handler", "HandleGeneralCallback")

	var callback models.CallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		logger.WithError(err).Error("failed to decode callback")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if callback.Transaction.Reference == "" {
		logger.Error("missing transaction reference in callback")
		http.Error(w, "Missing transaction reference", http.StatusBadRequest)
		return
	}

	logger = logger.WithFields(map[string]any{
		"transaction_ref": callback.Transaction.Reference,
		"callback_type":   callback.CallbackType,
		"amount":          callback.Transaction.Amount,
		"currency":        callback.Transaction.Currency,
		"status":          callback.Transaction.Status,
	})
	logger.Info("received general callback from Jenga")

	callbackJSON, _ := json.Marshal(callback)
	cbJSON := data.JSONMap{"additional_info": string(callbackJSON)}

	amtDec, _ := decimalx.NewFromString(fmt.Sprintf("%g", callback.Transaction.Amount))
	amount := utilmoney.ToMoney(callback.Transaction.Currency, amtDec)
	costDec, _ := decimalx.NewFromString(fmt.Sprintf("%g", callback.Transaction.ServiceCharge))
	cost := utilmoney.ToMoney(callback.Transaction.Currency, costDec)

	payment := &paymentv1.Payment{
		Source: &commonv1.ContactLink{
			Detail: callback.Customer.MobileNumber,
		},
		Recipient: &commonv1.ContactLink{
			Detail: callback.Bank.Account,
		},
		TransactionId: callback.Transaction.Reference,
		Amount:        amount,
		Cost:          cost,
		Extra:         cbJSON.ToProtoStruct(),
	}

	_, err := ws.paymentCli.Receive(ctx, connect.NewRequest(&paymentv1.ReceiveRequest{
		Data: payment,
	}))
	if err != nil {
		logger.WithError(err).Error("failed to forward callback to payment service")
		http.Error(w, "Failed to process callback", http.StatusInternalServerError)
		return
	}

	logger.Info("general callback processed successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck // best-effort response
		"status":  "success",
		"message": "Callback received successfully",
	})
}
