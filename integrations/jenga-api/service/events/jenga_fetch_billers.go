package events

import (
	"context"
	"fmt"

	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/go-redis/redis"
	"github.com/pitabwire/frame"
)

// JengaFetchBillers represents the event for fetching billers
// It includes the Redis client and service for event processing

type JengaFetchBillers struct {
	Service     *frame.Service
	Client      *coreapi.Client
	RedisClient *redis.Client
}

// Name returns the name of the event
func (event *JengaFetchBillers) Name() string {
	return "jenga.fetch.billers"
}

// PayloadType returns the type of the payload
func (event *JengaFetchBillers) PayloadType() any {
	return &models.FetchBillersRequest{}
}

// Validate validates the payload
func (event *JengaFetchBillers) Validate(ctx context.Context, payload any) error {
	// Add validation logic if needed
	return nil
}

// Execute handles the fetching of billers
func (event *JengaFetchBillers) Execute(ctx context.Context, payload any) error {
	//request := payload.(*models.FetchBillersRequest)

	// Fetch billers using the client
	billers, err := event.Client.FetchBillers()
	if err != nil {
		return fmt.Errorf("failed to fetch billers: %v", err)
	}

	// Log the successful fetching of billers
	event.Service.L(ctx).WithField("billers", billers).Info("Successfully fetched billers")

	return nil
}