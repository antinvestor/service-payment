package handlers

import (
	"encoding/json"
	"net/http"

	paymentv1connect "buf.build/gen/go/antinvestor/payment/connectrpc/go/payment/v1/paymentv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/coreapi"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/events/eventstillspay"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/pitabwire/frame/events"
	"github.com/pitabwire/util"
)

// job server handlers

type JobServer struct {
	eventMan      events.Manager
	client        coreapi.JengaApiClient
	paymentClient paymentv1connect.PaymentServiceClient
}

// NewJobServer creates a new JobServer with dependencies.
func NewJobServer(
	eventMan events.Manager,
	client coreapi.JengaApiClient,
	paymentClient paymentv1connect.PaymentServiceClient,
) *JobServer {
	return &JobServer{
		eventMan:      eventMan,
		client:        client,
		paymentClient: paymentClient,
	}
}

func (js *JobServer) InitiateTillsPay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// background context for async processing
	ctx := r.Context()
	logger := util.Log(ctx).WithField("type", "InitiateTillsPay")

	var request models.TillsPayRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.WithError(err).Error("failed to decode request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields (basic check, event will do full validation)
	if request.Merchant.Till == "" || request.Payment.Ref == "" || request.Payment.Amount == "" ||
		request.Payment.Currency == "" ||
		request.Partner.ID == "" ||
		request.Partner.Ref == "" {
		http.Error(w, "Invalid request: missing required fields", http.StatusBadRequest)
		return
	}

	// Create event handler
	event := eventstillspay.NewJengaTillsPay(js.client)

	// Execute event synchronously with request context
	err := js.eventMan.Emit(ctx, event.Name(), &request)
	if err != nil {
		logger.WithError(err).WithField("payment_ref", request.Payment.Ref).Error("failed to process tills pay request")
		http.Error(w, "Failed to process tills pay request", http.StatusInternalServerError)
		return
	}

	// Return success response after processing
	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(map[string]string{
		"status":      "success",
		"message":     "Tills pay request processed successfully",
		"referenceId": request.Payment.Ref,
	}); encodeErr != nil {
		logger.WithError(encodeErr).Error("failed to encode response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HealthHandler is a simple health check handler.
func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
