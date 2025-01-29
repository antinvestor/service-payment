package events

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/go-redis/redis"
	"github.com/pitabwire/frame"
)

type JengaGoodsServices struct {
	Service     *frame.Service
	Client      *coreapi.Client
	RedisClient *redis.Client
}

func (event *JengaGoodsServices) Name() string {
	return "jenga.goods.services"
}

func (event *JengaGoodsServices) PayloadType() any {
	return &models.Job{}
}

func (event *JengaGoodsServices) updateJobStatus(jobID string, status string, response interface{}) error {
	// Update status
	if err := event.RedisClient.Set(jobID+"_status", status, 0).Err(); err != nil {
		return errors.New("failed to update job status in redis")
	}

	// Store response if provided
	if response != nil {
		responseData, err := json.Marshal(response)
		if err != nil {
			return errors.New("failed to marshal response data")
		}

		if err := event.RedisClient.Set(jobID+"_response", string(responseData), 0).Err(); err != nil {
			return errors.New("failed to save response data to redis")
		}
	}

	return nil
}

func (event *JengaGoodsServices) Validate(ctx context.Context, payload any) error {
	job, ok := payload.(*models.Job)
	if !ok {
		return errors.New("payload is not of type models.Job")
	}

	request := job.ExtraData

	if request.Biller.BillerCode == "" {
		return errors.New("biller code is required")
	}

	if request.Bill.Amount == "" {
		return errors.New("bill amount is required")
	}

	if request.Bill.Reference == "" {
		return errors.New("bill reference is required")
	}

	if request.PartnerID == "" {
		return errors.New("partner ID is required")
	}

	return nil
}

func (event *JengaGoodsServices) Execute(ctx context.Context, payload any) error {
	job := payload.(*models.Job)
	request := job.ExtraData

	logger := event.Service.L(ctx).WithField("type", event.Name()).WithField("job_id", job.ID)
	logger.WithField("request", request).Debug("processing payment job")

	// Update status to processing
	if err := event.updateJobStatus(job.ID, "processing", nil); err != nil {
		logger.WithError(err).Error("failed to update job status to processing")
		return err
	}

	// Generate bearer token for authorization
	token, err := event.Client.GenerateBearerToken()
	if err != nil {
		logger.WithError(err).Error("failed to generate bearer token")
		event.updateJobStatus(job.ID, "failed", map[string]string{
			"error": "failed to generate bearer token",
		})
		return err
	}

	// Generate signature for the request
	signature := event.Client.GenerateSignatureBillGoodsAndServices(
		request.Biller.BillerCode,
		request.Bill.Amount,
		request.Bill.Reference,
		request.PartnerID,
	)

	// Add signature and job metadata to request
	metadata := map[string]string{
		"signature": signature,
		"job_id":    job.ID,
	}
	metadataBytes, _ := json.Marshal(metadata)
	request.Remarks = string(metadataBytes)

	// Initiate the payment
	response, err := event.Client.InitiateBillGoodsAndServices(request, token.AccessToken)
	if err != nil {
		logger.WithError(err).Error("failed to initiate payment")
		event.updateJobStatus(job.ID, "failed", map[string]string{
			"error": err.Error(),
		})
		return err
	}

	logger.WithField("response", response).Info("payment processed")

	// Handle the response based on status
	if !response.Status {
		event.updateJobStatus(job.ID, "failed", response)
		return errors.New(response.Message)
	}

	// Update final status with response
	if err := event.updateJobStatus(job.ID, "completed", response); err != nil {
		logger.WithError(err).Error("failed to update final status")
		return err
	}

	return nil
}
