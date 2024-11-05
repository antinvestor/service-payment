package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/antinvestor/service-payments/integrations/jenga-api/config"
	client "github.com/antinvestor/service-payments/integrations/jenga-api/service/coreapi"
)

func StkUssdPushHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var jengaConfig config.JengaConfig
	var request client.STKUSSDRequest
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
