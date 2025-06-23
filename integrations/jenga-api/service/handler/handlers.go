package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/events/events_account_balance"
	"github.com/antinvestor/jenga-api/service/events/events_stk"
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

func (js *JobServer) InitiateStkUssd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	logger := js.Service.Log(ctx).WithField("type", "InitiateStkUssd")

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
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":      "success",
		"message":     "STK/USSD push request accepted for processing",
		"referenceId": request.Payment.Ref, // Use the correct field Ref
	}); err != nil {
		logger.WithError(err).Error("failed to encode response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Process the request asynchronously in a goroutine
	go func(req models.STKUSSDRequest) {
		// Create a new background context for async processing
		bgCtx := context.Background()
		bgLogger := js.Service.Log(bgCtx).WithField("type", "AsyncInitiateStkUssd")
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

			statusUpdateRequest := &commonv1.StatusUpdateRequest{
				Id:     req.ID,
				State:  commonv1.STATE_ACTIVE,
				Status: commonv1.STATUS_FAILED,
			}
			_, err = js.PaymentClient.StatusUpdate(bgCtx, statusUpdateRequest)
			if err != nil {
				bgLogger.WithError(err).WithField("reference", req.Payment.Ref).Error("failed to update payment status")
			}
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

	logger := js.Service.Log(ctx).WithField("type", "AccountBalanceHandler")
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

	// Make a copy for async processing
	requestCopy := request

	// Return immediate success response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":      "success",
		"message":     "Tills pay request accepted for processing",
		"referenceId": request.Payment.Ref,
	}); err != nil {
		logger.WithError(err).Error("failed to encode response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Process the request asynchronously in a goroutine
	go func(req models.TillsPayRequest) {
		bgCtx := context.Background()
		bgLogger := js.Service.Log(bgCtx).WithField("type", "AsyncInitiateTillsPay")
		bgLogger.WithField("reference", req.Payment.Ref).Info("starting async tills pay processing")

		// Create event
		event := &events_tills_pay.JengaTillsPay{
			Service: js.Service,
			Client:  js.Client,
		}

		// Execute event
		err := js.Service.Emit(bgCtx, event.Name(), &req)
		if err != nil {
			bgLogger.WithError(err).WithField("reference", req.Payment.Ref).Error("failed to process async tills pay request")
			// Optionally: update status in external system if needed
		} else {
			bgLogger.WithField("reference", req.Payment.Ref).Info("successfully processed async tills pay request")
		}
	}(requestCopy)
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
