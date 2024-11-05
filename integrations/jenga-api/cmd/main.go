package cmd

import (
	"github.com/antinvestor/service-payments/integrations/jenga-api/service/router"
	"net/http"
)

//our main application

func main() {
	router := router.NewRouter()
	http.ListenAndServe(":8080", router)
}
