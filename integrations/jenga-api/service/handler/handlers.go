package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/events/events_stk"
	"github.com/antinvestor/jenga-api/service/events/events_account_balance"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/pitabwire/frame"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
)

//job server handlers

type JobServer struct {
	Service *frame.Service
	Client  *coreapi.Client
	PaymentClient *paymentV1.PaymentClient
}

func (js *JobServer) InitiateStkUssd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	logger := js.Service.L(ctx).WithField("type", "InitiateStkUssd")

	var request models.STKUSSDRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.WithError(err).Error("failed to decode request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the request by checking required fields
	if request.Merchant.AccountNumber == "" || request.Payment.MobileNumber == "" || request.Payment.Amount == "" {
		http.Error(w, "Invalid request: missing required fields", http.StatusBadRequest)
		return
	}

	// Make a copy of the request for async processing
	requestCopy := request

	// Return immediate success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "STK/USSD push request accepted for processing",
		"referenceId": request.Payment.Ref, // Use the correct field Ref
	})

	// Process the request asynchronously in a goroutine
	go func(req models.STKUSSDRequest) {
		// Create a new background context for async processing
		bgCtx := context.Background()
		bgLogger := js.Service.L(bgCtx).WithField("type", "AsyncInitiateStkUssd")
		bgLogger.WithField("reference", req.Payment.Ref).Info("starting async STK/USSD processing")

		// Create event
		event := &events_stk.JengaSTKUSSD{
			Service: js.Service,
			Client:  js.Client,
		}

		// Execute event
		err := js.Service.Emit(bgCtx, event.Name(), &req)
		if err != nil {
			bgLogger.WithError(err).WithField("reference", req.Payment.Ref).Error("failed to process async STK/USSD request")
			// Note: Cannot return HTTP error since we've already sent the response
			// Here you could implement a notification mechanism or fallback strategy
			//_ , err = js.PaymentClient.Client.Init

		} else {
			bgLogger.WithField("reference", req.Payment.Ref).Info("successfully processed async STK/USSD request")
		}
	}(requestCopy)
}

func (js *JobServer) AccountBalanceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	logger := js.Service.L(ctx).WithField("type", "AccountBalanceHandler")
	logger.Info("processing account balance")

	// https://uat.finserve.africa/v3-apis/account-api/v3.0/accounts/balances/{countryCode}/{accountId}
	//get the country code and account number from the request
	countryCode := r.URL.Query().Get("countryCode")
	accountNumber := r.URL.Query().Get("accountId")

	eventPayload := &models.AccountBalanceRequest{
		CountryCode: countryCode,
		AccountId:   accountNumber,
	}

	//processing event Payload
	logger.WithField("payload", eventPayload).Debug("------processing event-----------------------------------")

	event := &events_account_balance.JengaAccountBalance{

		Service: js.Service,
	}
	err := js.Service.Emit(ctx, event.Name(), eventPayload)
	if err != nil {
		logger.WithError(err).Error("failed to execute event")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HealthHandler is a simple health check handler
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

}
