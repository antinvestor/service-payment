package router

import (
	handlers "github.com/antinvestor/jenga-api/service/handler"
	"github.com/gorilla/mux"
)

func NewRouter(js *handlers.JobServer) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	// Health check endpoint
	router.HandleFunc("/health", handlers.HealthHandler).Methods("GET")
	// Callback endpoint
	router.HandleFunc("/receivepayments", js.HandleStkCallback).Methods("POST")
	router.HandleFunc("/payments/tills-pay", js.InitiateTillsPay).Methods("POST")
	return router
}
