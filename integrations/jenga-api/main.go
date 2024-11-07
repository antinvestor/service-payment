package main

import (
	"context"
	"github.com/antinvestor/jenga-api/service/router"
	"github.com/pitabwire/frame"
	"log"
)

//our main application

func main() {

	serviceName := "service_jenga_api"
	ctx := context.Background()
	router := router.NewRouter()

	server := frame.HttpHandler(router)

	ctx, service := frame.NewService(serviceName, server)
	err := service.Run(ctx, ":443")
	if err != nil {
		log.Fatal("main -- Could not run Server : %v", err)
	}

}
