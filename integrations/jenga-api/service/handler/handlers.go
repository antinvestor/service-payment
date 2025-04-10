package handlers

import (
	"encoding/json"
	"net/http"


	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/events"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/pitabwire/frame"
)

//job server handlers

type JobServer struct {
	Service     *frame.Service
	Client      *coreapi.Client
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

	// Create event
	event := &events.JengaSTKUSSD{
		Service: js.Service,
		Client:  js.Client,
	}

	// Execute event
	err := js.Service.Emit(ctx, event.Name(), &request)
	if err != nil {
		logger.WithError(err).Error("failed to process STK/USSD request")
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "STK/USSD push initiated successfully",
	})
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

	event := &events.JengaAccountBalance{
		
		Service:     js.Service,
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
