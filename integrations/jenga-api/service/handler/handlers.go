package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/antinvestor/jenga-api/service/coreapi"
	"github.com/antinvestor/jenga-api/service/events"
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/pitabwire/frame"
)

//job server handlers

type JobServer struct {
	Service     *frame.Service
	RedisClient *redis.Client
	Client       *coreapi.Client
}

func (js *JobServer) AsyncBillPaymentsGoodsandServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	logger := js.Service.L(ctx).WithField("type", "AsyncBillPaymentsGoodsandServices")

	var request models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.WithError(err).Error("failed to decode request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate job ID
	jobID := uuid.New().String()

	// Create job metadata
	metadata := map[string]interface{}{
		"job_id":        jobID,
		"request_type":  "jenga.goods.services",
		"created_at":    time.Now().UTC().Format(time.RFC3339),
		"initial_state": "queued",
	}

	// Store initial job state in Redis
	if err := js.RedisClient.Set(jobID+"_status", "queued", 24*time.Hour).Err(); err != nil {
		logger.WithError(err).Error("failed to set initial job status")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store request data in Redis
	requestData, _ := json.Marshal(request)
	if err := js.RedisClient.Set(jobID+"_request", string(requestData), 24*time.Hour).Err(); err != nil {
		logger.WithError(err).Error("failed to store request data")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Emit event for processing
	event := events.JengaGoodsServices{
		Service:     js.Service,
		RedisClient: js.RedisClient,
	}
	eventPayload := &models.Job{
		ID:        jobID,
		ExtraData: request, // request is already a models.PaymentRequest
	}

	if err := js.Service.Emit(ctx, event.Name(), eventPayload); err != nil {
		logger.WithError(err).Error("failed to emit event")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return job ID to client
	response := map[string]interface{}{
		"job_id":     jobID,
		"status":     "queued",
		"message":    "Payment request queued for processing",
		"created_at": metadata["created_at"],
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

func (js *JobServer) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	logger := js.Service.L(ctx).WithField("type", "GetJobStatus")

	jobID := r.URL.Query().Get("job_id")
	if jobID == "" {
		http.Error(w, "job_id is required", http.StatusBadRequest)
		return
	}

	// Get status from Redis
	status, err := js.RedisClient.Get(jobID + "_status").Result()
	if err == redis.Nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	} else if err != nil {
		logger.WithError(err).Error("failed to get job status")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get request data if available
	requestData, _ := js.RedisClient.Get(jobID + "_request").Result()

	// Get response data if available
	responseData, _ := js.RedisClient.Get(jobID + "_response").Result()

	response := map[string]interface{}{
		"job_id": jobID,
		"status": status,
	}

	if requestData != "" {
		var req interface{}
		json.Unmarshal([]byte(requestData), &req)
		response["request"] = req
	}

	if responseData != "" {
		var resp interface{}
		json.Unmarshal([]byte(responseData), &resp)
		response["response"] = resp
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (js *JobServer) AccountBalanceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	logger := js.Service.L(ctx).WithField("type", "AccountBalanceHandler")
	logger.Info("processing account balance")

	// https://uat.finserve.africa/v3-apis/account-api/v3.0/accounts/balances/{countryCode}/{accountId}
	//get the country code and account number from the request
	countryCode := r.URL.Query().Get("countryCode")
	accountNumber := r.URL.Query().Get("accountId")

	eventPayload := &models.AccountBalanceRequest{
		CountryCode: countryCode,
		AccountId:   accountNumber,
	}

	//processing event Payload
	logger.WithField("payload", eventPayload).Debug("------processing event-----------------------------------")

	event := &events.JengaAccountBalance{
		RedisClient: js.RedisClient,
		Service:     js.Service,
	}
	err := js.Service.Emit(ctx, event.Name(), eventPayload)
	if err != nil {
		logger.WithError(err).Error("failed to execute event")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// FetchBillersHandler handles requests to fetch billers
func (js *JobServer) FetchBillersHandler(w http.ResponseWriter, r *http.Request) {
	// Log the request event
	logger := js.Service.L(r.Context()).WithField("type", "FetchBillers")

	// Log the request
	logger.Info("processing fetch billers")

	// Prepare event payload
	eventPayload := &models.FetchBillersRequest{}

	// Log the processing of event payload
	logger.WithField("payload", eventPayload).Debug("processing fetch billers event")

	// Create event
	event := &events.JengaFetchBillers{
		RedisClient: js.RedisClient,
		Service:     js.Service,
	}

	// Emit event
	if err := js.Service.Emit(r.Context(), event.Name(), eventPayload); err != nil {
		logger.WithError(err).Error("failed to execute fetch billers event")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HealthHandler is a simple health check handler
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

}
