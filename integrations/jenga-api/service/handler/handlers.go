package handlers

import (
	"encoding/json"
	"net/http"

	client "github.com/antinvestor/service-payments-v1/integrations/jenga-api/service/coreapi"
	
)

// HandleSTKPush handles STK push requests
func StkUssdPushHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var request client.STKUSSDRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	//initialize the client
	client := client.New("merchantCode", "consumerSecret", "apiKey", "env")
	//initiate STK/USSD push request
	response, err := client.InitiateSTKUSSD(request, "accessToken")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

