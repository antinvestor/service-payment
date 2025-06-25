package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/antinvestor/jenga-api/service/models"
)

func (js *JobServer) HandleStkCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := js.Service.Log(ctx).WithField("type", "CallbackHandler")

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Verify Basic Auth if needed
	// authHeader := r.Header.Get("Authorization")
	// if authHeader == "" {
	// 	http.Error(w, "Authorization header required", http.StatusUnauthorized)
	// 	return
	// }

	var callback models.StkCallback
	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		logger.WithError(err).Error("failed to decode callback request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields in the callback
	if callback.Transaction == "" || callback.MobileNumber == "" || callback.Currency == "" {
		logger.Error("missing required fields in callback")
		http.Error(w, "Missing required fields in callback", http.StatusBadRequest)
		return
	} 

	// Log the callback for debugging
	logger.WithField("callback", callback).Info("received callback")

	// Add additional information to the callback context for logging
	logger = logger.
		WithField("transaction_ref", callback.Transaction).
		WithField("telco_ref", callback.Telco).
		WithField("status", callback.Status).
		WithField("mobile_number", callback.MobileNumber)

	// Process the callback synchronously using the request's context
	err := js.Service.Emit(ctx, "jenga.callback.receive.payment", &callback)
	if err != nil {
		logger.WithError(err).Error("failed to emit callback event")
		http.Error(w, "Failed to process callback", http.StatusInternalServerError)
		return
	}

	logger.Info("Callback processed successfully")

	// Return success response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "Callback received successfully",
	}); err != nil {
		logger.WithError(err).Error("failed to encode success response")
	}
}
