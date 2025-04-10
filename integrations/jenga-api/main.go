package main

import (
	
	"fmt"
	"log"
	"os"

	commonV1 "github.com/antinvestor/apis/go/common"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	handler "github.com/antinvestor/jenga-api/service/handler"
	"github.com/antinvestor/jenga-api/config"
	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/events"
	"github.com/antinvestor/jenga-api/service/router"
	"github.com/go-redis/redis"
	"github.com/pitabwire/frame"
)

func main() {
	// Set config file path
	if err := os.Setenv("CONFIG_FILE", "config.yaml"); err != nil {
		log.Fatalf("failed to set config file env: %v", err)
	}

	serviceName := "service_jenga_api"
	var jengaConfig config.JengaConfig
	err := frame.ConfigProcess("", &jengaConfig)
	if err != nil {
		log.Fatalf("failed to process config: %v", err)
		return
	}
	//initialize jenga client
	if jengaConfig.MerchantCode == "" {
		log.Fatalf("MerchantCode is required")
		return
	}
	clientApi := coreapi.New(jengaConfig.MerchantCode, jengaConfig.ConsumerSecret, jengaConfig.ApiKey, jengaConfig.Env,jengaConfig.JengaPrivateKey)

	// Initialize payment client
    ctx , service := frame.NewService(serviceName, frame.Config(&jengaConfig))
	defer service.Stop(ctx)
	clientBase, err := commonV1.NewClientBase(ctx, commonV1.WithEndpoint("payment_service:50051"))
	if err != nil {
		log.Fatalf("failed to create client base: %v", err)
	}
	
	paymentClient := paymentV1.Init(clientBase, paymentV1.NewPaymentServiceClient(clientBase.Connection()))
	

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", jengaConfig.RedisHost, jengaConfig.RedisPort),	
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Initialize JobServer
	js := &handler.JobServer{
		Service:     service,
		RedisClient: redisClient,
	}

	// Initialize router
	router := router.NewRouter(js)

	// Create service options
	serviceOptions := []frame.Option{
		frame.HttpHandler(router),
		frame.RegisterEvents(
			&events.JengaGoodsServices{Service: service, RedisClient: redisClient, Client: clientApi},
			&events.JengaAccountBalance{Service: service, RedisClient: redisClient, Client: clientApi},
			&events.JengaCallbackReceivePayment{Service: service, PaymentClient: paymentClient},
			&events.JengaFetchBillers{Service: service, RedisClient: redisClient, Client: clientApi},
		),
	}

	// Start the service
	service.Init(serviceOptions...)

	log.Printf("Starting Jenga API service on :8080")
	err = service.Run(ctx, "")
	if err != nil {
		log.Fatalf("failed to run service: %v", err)
	}

	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("error closing redis client: %v", err)
		}
	}()
}
