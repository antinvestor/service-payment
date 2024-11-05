package router

//use the mux router
import (
	handlers "github.com/antinvestor/service-payments/integrations/jenga-api/service/handler"
	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/stk-ussd-push", handlers.StkUssdPushHandler).Methods("POST")
	return router
}
