package main

import (
	"fmt"
	"log"
	"os"

	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/config"
	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/events"
	handler "github.com/antinvestor/jenga-api/service/handler"
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
	clientApi := coreapi.New(jengaConfig.MerchantCode, jengaConfig.ConsumerSecret, jengaConfig.ApiKey, jengaConfig.Env, jengaConfig.JengaPrivateKey)

	// Initialize payment client
	//paymentConn, err := grpc.Dial(jengaConfig.ProfileServiceURI, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to payment service: %v", err)
		return
	}


	paymentConn, err := grpc.Dial("payment_service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to payment service: %v", err)
		return
	}
	
	paymentClient := paymentV1.NewPaymentsClient(paymentConn)
	// Get Redis configuration from environment
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
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
