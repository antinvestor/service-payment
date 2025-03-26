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
	
	router.HandleFunc("/payments/goods-services",js.AsyncBillPaymentsGoodsandServices).Methods("POST")
	router.HandleFunc("/jobs/{jobID}", js.GetJobStatus).Methods("GET")
	router.HandleFunc("/account-balance", js.AccountBalanceHandler).Methods("GET")
	//get billers
	
	
	// Callback endpoint
	router.HandleFunc("/receivepayments", js.HandleCallback).Methods("POST")

	return router
}