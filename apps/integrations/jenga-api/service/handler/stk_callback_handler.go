package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/pitabwire/util"
)

func (js *JobServer) HandleStkCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := util.Log(ctx).WithField("handler", "HandleStkCallback")

	var callback models.StkCallback
	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		logger.WithError(err).Error("failed to decode callback request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if callback.Transaction == "" || callback.MobileNumber == "" || callback.Currency == "" {
		logger.Error("missing required fields in callback")
		http.Error(w, "Missing required fields in callback", http.StatusBadRequest)
		return
	}

	logger = logger.WithFields(map[string]any{
		"transaction_ref": callback.Transaction,
		"mobile_number":   callback.MobileNumber,
		"status":          callback.Status,
		"amount":          callback.RequestAmount,
		"currency":        callback.Currency,
	})
	logger.Info("received STK callback")

	err := js.eventMan.Emit(ctx, "jenga.callback.receive.payment", &callback)
	if err != nil {
		logger.WithError(err).Error("failed to emit callback event")
		http.Error(w, "Failed to process callback", http.StatusInternalServerError)
		return
	}

	logger.Info("STK callback processed successfully")

	w.WriteHeader(http.StatusOK)
	if encodeErr := json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Callback received successfully",
	}); encodeErr != nil {
		logger.WithError(encodeErr).Error("failed to encode success response")
	}
}
