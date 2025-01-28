package worker

import (
	"context"
	"encoding/json"
	"github.com/antinvestor/jenga-api/config"
	client "github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/go-redis/redis"
	"github.com/pitabwire/frame"
	"net/http"
)

type Workers struct {
	Service     *frame.Service
	redisClient *redis.Client
}

func (ws *Workers) BillPaymentsGoodsandServices(ctx context.Context, job models.Job, w http.ResponseWriter, r *http.Request) {
	ws.Service.L(ctx).Info("BillPaymentsGoodsandServices started")
	result := job.ExtraData
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := ws.redisClient.Set(job.ID+"_status", "pending", 0).Err()
	if err != nil {
		http.Error(w, "could not save job to redis", http.StatusInternalServerError)
		return
	}
	err = ws.redisClient.Set(job.ID, result, 0).Err()
	if err != nil {
		http.Error(w, "could not save job to redis", http.StatusInternalServerError)
		return
	}
	var jengaConfig config.JengaConfig
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
	//initiate bill goods and services
	response, err := clientApi.InitiateBillGoodsAndServices(result, bearerToken.AccessToken)

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
