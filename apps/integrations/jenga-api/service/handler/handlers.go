package handlers

import (
	"encoding/json"
	"net/http"

	paymentv1connect "buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/coreapi"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/events/eventstillspay"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/pitabwire/frame/events"
	"github.com/pitabwire/util"
)

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

// NewRouter creates the HTTP router with all Jenga API routes.
func (js *JobServer) NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /receivepayments", js.HandleStkCallback)
	mux.HandleFunc("POST /payments/tills-pay", js.InitiateTillsPay)
	return mux
}

func (js *JobServer) InitiateTillsPay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := util.Log(ctx).WithField("handler", "InitiateTillsPay")

	var request models.TillsPayRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.WithError(err).Error("failed to decode request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Merchant.Till == "" || request.Payment.Ref == "" || request.Payment.Amount == "" ||
		request.Payment.Currency == "" ||
		request.Partner.ID == "" ||
		request.Partner.Ref == "" {
		logger.Error("missing required fields in tills pay request")
		http.Error(w, "Invalid request: missing required fields", http.StatusBadRequest)
		return
	}

	logger = logger.WithFields(map[string]any{
		"payment_ref": request.Payment.Ref,
		"till":        request.Merchant.Till,
		"amount":      request.Payment.Amount,
		"currency":    request.Payment.Currency,
	})
	logger.Info("processing tills pay request")

	event := eventstillspay.NewJengaTillsPay(js.client)

	err := js.eventMan.Emit(ctx, event.Name(), &request)
	if err != nil {
		logger.WithError(err).Error("failed to process tills pay request")
		http.Error(w, "Failed to process tills pay request", http.StatusInternalServerError)
		return
	}

	logger.Info("tills pay request processed successfully")

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
