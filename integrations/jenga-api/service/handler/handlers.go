package handlers

import (
	"context"
	"encoding/json"
	"github.com/antinvestor/jenga-api/config"
	client "github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/pitabwire/frame"
	"net/http"
)

//job server handlers

type JobServer struct {
	service *frame.Service
	redis   *redis.Client
}

func (js *JobServer) asyncBillPaymentsGoodsandServices(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	js.service.L(ctx).Info("asyncBillPaymentsGoodsandServices started")
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	//var jengaConfig config.JengaConfig
	var request models.PaymentRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	jobID := uuid.New().String()
	js.service.L(ctx).Info("created new job with id: ", jobID)
	jsonBody, err := json.Marshal(models.Job{ID: jobID,
		Type:      "asyncBillPaymentsGoodsandServices",
		ExtraData: request,
	})

	err = js.service.Publish(ctx, jobID, jsonBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]string{"status": "ok", "job_id": jobID}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	js.service.L(ctx).Info("asyncBillPaymentsGoodsandServices completed")
}

// HealthHandler is a simple health check handler
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

}

func StkUssdPushHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var jengaConfig config.JengaConfig
	var request models.STKUSSDRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	clientApi := client.New(jengaConfig.MerchantCode, jengaConfig.ConsumerSecret, jengaConfig.Env, jengaConfig.ApiKey)

	//generate bearer token
	var bearerToken *client.BearerTokenResponse
	bearerToken, err = clientApi.GenerateBearerToken()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//initiate STK/USSD push request
	response, err := clientApi.InitiateSTKUSSD(request, bearerToken.AccessToken)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	//test error after encoding
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
