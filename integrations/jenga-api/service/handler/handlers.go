package handlers

import (
	"encoding/json"
	"net/http"

	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/events/events_tills_pay"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/pitabwire/frame"
)

//job server handlers

type JobServer struct {
	Service       *frame.Service
	Client        *coreapi.Client
	PaymentClient *paymentV1.PaymentClient
}



func (js *JobServer) InitiateTillsPay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	//background context for async processing
	ctx := r.Context()
	logger := js.Service.Log(ctx).WithField("type", "InitiateTillsPay")

	var request models.TillsPayRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.WithError(err).Error("failed to decode request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields (basic check, event will do full validation)
	if request.Merchant.Till == "" || request.Payment.Ref == "" || request.Payment.Amount == "" || request.Payment.Currency == "" || request.Partner.ID == "" || request.Partner.Ref == "" {
		http.Error(w, "Invalid request: missing required fields", http.StatusBadRequest)
		return
	}

	// Create event
	event := &events_tills_pay.JengaTillsPay{
		Service: js.Service,
		Client:  js.Client,
	}

	// Execute event synchronously with request context
	err := js.Service.Emit(ctx, event.Name(), &request)
	if err != nil {
		logger.WithError(err).WithField("reference", request.Payment.Ref).Error("failed to process tills pay request")
		http.Error(w, "Failed to process tills pay request", http.StatusInternalServerError)
		return
	}

	// Return success response after processing
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":      "success",
		"message":     "Tills pay request processed successfully",
		"referenceId": request.Payment.Ref,
	}); err != nil {
		logger.WithError(err).Error("failed to encode response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HealthHandler is a simple health check handler.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
