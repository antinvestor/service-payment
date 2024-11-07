package router

//use the mux router
import (
	handlers "github.com/antinvestor/jenga-api/service/handler"
	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	//for testing purposes
	router.HandleFunc("/health", handlers.HealthHandler).Methods("GET")
	router.HandleFunc("/stk-ussd-push", handlers.StkUssdPushHandler).Methods("POST")
	return router
}
