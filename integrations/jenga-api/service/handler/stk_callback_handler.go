package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/antinvestor/jenga-api/service/models"
)

func (js *JobServer) HandleStkCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	logger := js.Service.L(ctx).WithField("type", "CallbackHandler")

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
	logger = logger.WithFields(map[string]interface{}{
		"transaction_ref": callback.Transaction,
		"telco_ref": callback.Telco,
		"status": callback.Status,
	})
	
	// Create a background context for the goroutine that won't be canceled when the request ends
	// Copy any relevant values from the request context
	bgCtx := context.Background()
	
	// Queue the callback for processing using the event system in a goroutine
	go func(callbackData models.StkCallback) {
		// Use a separate logger for the goroutine to avoid race conditions
		gLogger := js.Service.L(bgCtx).WithField("type", "CallbackProcessing")
		
		// Add additional information to the callback context for logging
		gLogger = gLogger.WithFields(map[string]interface{}{
			"transaction_ref": callbackData.Transaction,
			"telco_ref": callbackData.Telco,
			"status": callbackData.Status,
		})
		
		err := js.Service.Emit(bgCtx, "jenga.callback.receive.payment", &callbackData)
		if err != nil {
			gLogger.WithError(err).Error("failed to emit callback event in background processor")
			return
		}
		gLogger.Info("Successfully processed callback event in background")
	}(callback) // Pass callback by value to avoid race conditions
	
	logger.Info("Callback accepted for processing")

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "Callback received successfully",
	})
}
