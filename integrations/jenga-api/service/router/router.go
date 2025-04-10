package router

import (
	handlers "github.com/antinvestor/jenga-api/service/handler"
	"github.com/gorilla/mux"
)

func NewRouter(js *handlers.JobServer) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// Health check endpoint
	router.HandleFunc("/health", handlers.HealthHandler).Methods("GET")

   // Job related endpoints
	router.HandleFunc("/payments/stk-ussd", js.InitiateStkUssd).Methods("POST")
	router.HandleFunc("/account-balance", js.AccountBalanceHandler).Methods("GET")
	//get billers
	
	
	// Callback endpoint
	router.HandleFunc("/receivepayments", js.HandleCallback).Methods("POST")

	return router
}
