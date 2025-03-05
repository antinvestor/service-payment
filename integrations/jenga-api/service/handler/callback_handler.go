package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/antinvestor/jenga-api/service/models"
)

func (js *JobServer) HandleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	logger := js.Service.L(ctx).WithField("type", "CallbackHandler")

	// Verify Basic Auth if needed
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	var callback models.CallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		logger.WithError(err).Error("failed to decode callback request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Log the callback for debugging
	logger.WithField("callback", callback).Info("received callback")

	callbackData, err := json.Marshal(callback)
	if err != nil {
		logger.WithError(err).Error("failed to marshal callback data")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Queue the callback for processing
	if err := js.Service.Publish(ctx, "jenga.callback.receive.payment", &callbackData); err != nil {
		logger.WithError(err).Error("failed to queue callback for processing")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusAccepted)
	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "Callback received successfully",
	})
}
