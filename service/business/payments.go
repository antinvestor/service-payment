package business

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/antinvestor/service-payments/service/repository"
	"google.golang.org/genproto/googleapis/type/money"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	partitionV1 "github.com/antinvestor/apis/go/partition/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments/service/events"
	"github.com/antinvestor/service-payments/service/models"
	"github.com/pitabwire/frame"
	"github.com/shopspring/decimal"
)

type PaymentBusiness interface {
	Send(ctx context.Context, payment *paymentV1.Payment) (*commonv1.StatusResponse, error)
	Receive(ctx context.Context, payment *paymentV1.Payment) (*commonv1.StatusResponse, error)
	Status(ctx context.Context, status *commonv1.StatusRequest) (*commonv1.StatusResponse, error)
	StatusUpdate(ctx context.Context, req *commonv1.StatusUpdateRequest) (*commonv1.StatusResponse, error)
	Release(ctx context.Context, status *paymentV1.ReleaseRequest) (*commonv1.StatusResponse, error)
	Search(search *commonv1.SearchRequest, stream paymentV1.PaymentService_SearchServer) error
	InitiatePrompt(ctx context.Context, req *paymentV1.InitiatePromptRequest) (*commonv1.StatusResponse, error)
}

func NewPaymentBusiness(_ context.Context, service *frame.Service, profileCli *profileV1.ProfileClient, partitionCli *partitionV1.PartitionClient) (PaymentBusiness, error) {
	//initialize the service
	if service == nil || profileCli == nil || partitionCli == nil {
		return nil, ErrorInitializationFail
	}
	return &paymentBusiness{
		service:      service,
		profileCli:   profileCli,
		partitionCli: partitionCli,
	}, nil
}

type paymentBusiness struct {
	service      *frame.Service
	profileCli   *profileV1.ProfileClient
	partitionCli *partitionV1.PartitionClient
}

func (pb *paymentBusiness) Send(ctx context.Context, message *paymentV1.Payment) (*commonv1.StatusResponse, error) {
	//logger := pb.service.L(ctx).WithField("request", message)

	//authClaim := frame.ClaimsFromContext(ctx)

	//logger.WithField("auth claim", authClaim).Info("handling queue out request")

	// Initialize Payment model
	p := &models.Payment{
		SenderProfileType:    message.GetSource().GetProfileType(),
		SenderProfileID:      message.GetSource().GetProfileId(),
		SenderContactID:      message.GetSource().GetContactId(),
		RecipientProfileType: message.GetRecipient().GetProfileType(),
		RecipientProfileID:   message.GetRecipient().GetProfileId(),
		RecipientContactID:   message.GetRecipient().GetContactId(),
		ReferenceId:          message.GetReferenceId(),
		BatchId:              message.GetBatchId(),
		RouteID:              message.GetRoute(),
		PaymentType:          "Bank Transfers",
		OutBound:             true,
	}

	//initialize cost

	c := &models.Cost{
		Amount: decimal.NullDecimal{
			Valid:   true,
			Decimal: decimal.NewFromFloat(float64(message.GetCost().Units)),
		},
		Currency: message.GetCost().CurrencyCode,
	}

	p.Cost = c

	// Generate or validate Payment ID
	if message.GetId() == "" {
		p.GenID(ctx)
	}

	pb.validateAmountAndCost(message, p, c)

	pStatus := models.PaymentStatus{
		PaymentID: message.GetId(),
		State:     int32(commonv1.STATE_CREATED.Number()),
		Status:    int32(commonv1.STATUS_QUEUED.Number()),
	}

	// Emit events for Payment and PaymentStatus
	event := events.PaymentSave{}
	if err := pb.service.Emit(ctx, event.Name(), p); err != nil {
		pb.service.L(ctx).WithError(err).Warn("could not emit payment event")
		return nil, err
	}

	eventStatus := events.PaymentStatusSave{}
	if err := pb.service.Emit(ctx, eventStatus.Name(), pStatus); err != nil {
		pb.service.L(ctx).WithError(err).Warn("could not emit payment status event")
		return nil, err
	}
	return pStatus.ToStatusAPI(), nil
}

func (pb *paymentBusiness) Receive(ctx context.Context, message *paymentV1.Payment) (*commonv1.StatusResponse, error) {
	logger := pb.service.L(ctx).WithField("request", message)
	logger.Info("handling receive request")
	//authClaim := frame.ClaimsFromContext(ctx)
	//logger.WithField("auth claim", authClaim).Info("handling send request")

	p := &models.Payment{
		SenderProfileType:    message.GetSource().GetProfileType(),
		SenderProfileID:      message.GetSource().GetProfileId(),
		SenderContactID:      message.GetSource().GetContactId(),
		RecipientProfileType: message.GetRecipient().GetProfileType(),
		RecipientProfileID:   message.GetRecipient().GetProfileId(),
		RecipientContactID:   message.GetRecipient().GetContactId(),
		ReferenceId:          message.GetReferenceId(),
		BatchId:              message.GetBatchId(),
		RouteID:              message.GetRoute(),
		OutBound:             false,
	}

	c := &models.Cost{
		Amount: decimal.NullDecimal{
			Valid:   true,
			Decimal: decimal.NewFromFloat(float64(message.GetCost().Units)),
		},
		Currency: message.GetCost().CurrencyCode,
	}

	p.Cost = c

	// Generate or validate Payment ID
	if message.GetId() == "" {
		p.GenID(ctx)
	}
	pb.validateAmountAndCost(message, p, c)

	// Set initial PaymentStatus
	pStatus := models.PaymentStatus{
		PaymentID: message.GetId(),
		State:     int32(commonv1.STATE_CREATED.Number()),
		Status:    int32(commonv1.STATUS_QUEUED.Number()),
	}

	event := events.PaymentSave{}
	if err := pb.service.Emit(ctx, event.Name(), p); err != nil {
		pb.service.L(ctx).WithError(err).Warn("could not emit payment event")
		return nil, err
	}

	eventStatus := events.PaymentStatusSave{}
	if err := pb.service.Emit(ctx, eventStatus.Name(), pStatus); err != nil {
		pb.service.L(ctx).WithError(err).Warn("could not emit payment status event")
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}

func (pb *paymentBusiness) Status(ctx context.Context, status *commonv1.StatusRequest) (*commonv1.StatusResponse, error) {
	logger := pb.service.L(ctx).WithField("request", status)
	logger.Info("handling status check request")

	// Define a slice of status handlers that will be tried in order
	statusHandlers := []func(context.Context, string) (*commonv1.StatusResponse, error){
		pb.getPaymentStatus,
		pb.getPromptStatus,
		// Add new handlers here in the future
	}

	// Try each handler until one succeeds
	for _, handler := range statusHandlers {
		statusResponse, err := handler(ctx, status.GetId())
		if err == nil {
			// Successfully found and processed the status
			return statusResponse, nil
		}
		// Log but continue to next handler
		logger.WithError(err).Debug("status handler couldn't process request")
	}

	// If we get here, no handler could process the request
	logger.Error("could not find entity with this ID")
	return nil, fmt.Errorf("no entity found with ID: %s", status.GetId())
}

func (pb *paymentBusiness) StatusUpdate(ctx context.Context, req *commonv1.StatusUpdateRequest) (*commonv1.StatusResponse, error) {
	logger := pb.service.L(ctx).WithField("request", req)
	logger.Info("handling status update request")

	// Define a slice of status update handlers that will be tried in order
	statusUpdateHandlers := []func(context.Context, *commonv1.StatusUpdateRequest) (*commonv1.StatusResponse, error){
		pb.updatePaymentStatus,
		pb.updatePromptStatus,
		// Add new handlers here in the future
	}

	// Try each handler until one succeeds
	for _, handler := range statusUpdateHandlers {
		statusResponse, err := handler(ctx, req)
		if err == nil {
			// Successfully updated the status
			return statusResponse, nil
		}
		// Log but continue to next handler
		logger.WithError(err).Debug("status update handler couldn't process request")
	}

	// If we get here, no handler could process the request
	logger.Error("could not find entity with this ID")
	return nil, fmt.Errorf("no entity found with ID: %s", req.GetId())
}

// getPaymentStatus tries to get the status of a payment with the given ID
func (pb *paymentBusiness) getPaymentStatus(ctx context.Context, id string) (*commonv1.StatusResponse, error) {
	logger := pb.service.L(ctx).WithField("paymentId", id)

	// Try to find a payment with this ID
	paymentRepo := repository.NewPaymentRepository(ctx, pb.service)
	p, err := paymentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get the payment status
	paymentStatusRepo := repository.NewPaymentStatusRepository(ctx, pb.service)
	pStatus, err := paymentStatusRepo.GetByID(ctx, p.ID)
	if err != nil {
		logger.WithError(err).Error("could not get payment status")
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}

// getPromptStatus tries to get the status of a prompt with the given ID
func (pb *paymentBusiness) getPromptStatus(ctx context.Context, id string) (*commonv1.StatusResponse, error) {
	logger := pb.service.L(ctx).WithField("promptId", id)

	// Try to find a prompt with this ID
	promptRepo := repository.NewPromptRepository(ctx, pb.service)
	prompt, err := promptRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get the prompt status
	promptStatusRepo := repository.NewPromptStatusRepository(ctx, pb.service)
	pStatus, err := promptStatusRepo.GetByID(ctx, prompt.ID)
	if err != nil {
		logger.WithError(err).Error("could not get prompt status")
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}

// updatePaymentStatus tries to update the status of a payment with the given ID
func (pb *paymentBusiness) updatePaymentStatus(ctx context.Context, req *commonv1.StatusUpdateRequest) (*commonv1.StatusResponse, error) {
	logger := pb.service.L(ctx).WithField("paymentId", req.GetId())

	// Try to find a payment with this ID
	paymentRepo := repository.NewPaymentRepository(ctx, pb.service)
	p, err := paymentRepo.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	// Create a payment status update
	pStatus := models.PaymentStatus{
		PaymentID: p.ID,
		State:     int32(req.GetState()),
		Status:    int32(req.GetStatus()),
		Extra:     frame.DBPropertiesFromMap(req.GetExtras()),
	}

	pStatus.GenID(ctx)

	// Emit the payment status event
	eventStatus := events.PaymentStatusSave{}
	err = pb.service.Emit(ctx, eventStatus.Name(), pStatus)
	if err != nil {
		logger.WithError(err).Warn("could not save payment status")
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}

// updatePromptStatus tries to update the status of a prompt with the given ID
func (pb *paymentBusiness) updatePromptStatus(ctx context.Context, req *commonv1.StatusUpdateRequest) (*commonv1.StatusResponse, error) {
	logger := pb.service.L(ctx).WithField("promptId", req.GetId())

	// Try to find a prompt with this ID
	promptRepo := repository.NewPromptRepository(ctx, pb.service)
	prompt, err := promptRepo.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	// Create a prompt status update
	pStatus := models.PromptStatus{
		PromptID: prompt.ID,
		State:    int32(req.GetState()),
		Status:   int32(req.GetStatus()),
		Extra:    frame.DBPropertiesFromMap(req.GetExtras()),
	}

	pStatus.GenID(ctx)

	// Emit the prompt status event
	eventStatus := events.PromptStatusSave{}
	err = pb.service.Emit(ctx, eventStatus.Name(), pStatus)
	if err != nil {
		logger.WithError(err).Error("could not emit prompt status event")
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}

func (pb *paymentBusiness) Search(search *commonv1.SearchRequest,
	stream paymentV1.PaymentService_SearchServer) error {
	logger := pb.service.L(stream.Context()).WithField("request", search)
	logger.Debug("handling payment search request")

	// Extract the context and JWT token
	ctx := stream.Context()
	jwtToken := frame.JwtFromContext(ctx)
	logger.WithField("jwt", jwtToken).Debug("auth jwt supplied")

	// Initialize the payment repository
	paymentRepo := repository.NewPaymentRepository(ctx, pb.service)

	var paymentList []*models.Payment
	var err error

	// Handle search by ID or by general query
	if search.GetIdQuery() != "" {
		// Search by ID
		payment, err0 := paymentRepo.GetByID(ctx, search.GetIdQuery())
		if err0 != nil {
			return err0
		}

		paymentList = append(paymentList, payment)
	} else {
		// General search query
		paymentList, err = paymentRepo.Search(ctx, search.GetQuery())
		if err != nil {
			logger.WithError(err).Error("failed to search payments")
			return err
		}
	}

	// Initialize the payment status repository
	paymentStatusRepo := repository.NewPaymentStatusRepository(ctx, pb.service)

	var resultStatus *models.PaymentStatus
	var responsesList []*paymentV1.Payment
	for _, p := range paymentList {
		pStatus := &models.PaymentStatus{}
		if p.ID != "" {
			resultStatus, err = paymentStatusRepo.GetByID(ctx, p.ID)
			if err != nil {
				logger.WithError(err).WithField("status_id", p.ID).Error("could not get status id")
				return err
			} else {
				pStatus = resultStatus
			}
		}

		// Convert the payment model to the API response format
		result := p.ToApi(pStatus, nil)
		responsesList = append(responsesList, result)
	}

	// Send the search response back to the client
	err = stream.Send(&paymentV1.SearchResponse{Data: responsesList})
	if err != nil {
		logger.WithError(err).Warn("unable to send a result")
	}

	return nil
}

func (pb *paymentBusiness) Release(ctx context.Context, paymentReq *paymentV1.ReleaseRequest) (*commonv1.StatusResponse, error) {
	logger := pb.service.L(ctx).WithField("request", paymentReq)
	logger.Debug("handling release request")

	paymentRepo := repository.NewPaymentRepository(ctx, pb.service)
	p, err := paymentRepo.GetByID(ctx, paymentReq.GetId())
	if err != nil {
		logger.WithError(err).Warn("could not fetch payment by id")
		return nil, err
	}

	if !p.IsReleased() {
		releaseDate := time.Now()
		p.ReleasedAt = &releaseDate

		event := events.PaymentSave{}
		err = pb.service.Emit(ctx, event.Name(), p)
		if err != nil {
			logger.WithError(err).Warn("could not emit payment save")
			return nil, err
		}

		pStatus := models.PaymentStatus{
			PaymentID: p.GetID(),
			State:     int32(commonv1.STATE_ACTIVE.Number()),
			Status:    int32(commonv1.STATUS_QUEUED.Number()),
		}

		pStatus.GenID(ctx)

		eventStatus := events.PaymentStatusSave{}
		err = pb.service.Emit(ctx, eventStatus.Name(), pStatus)
		if err != nil {
			logger.WithError(err).Warn("could not emit payment status")
			return nil, err
		}

		return pStatus.ToStatusAPI(), nil
	} else {
		paymentStatusRepo := repository.NewPaymentStatusRepository(ctx, pb.service)
		pStatus, err := paymentStatusRepo.GetByID(ctx, p.ID)
		if err != nil {
			logger.WithError(err).Warn("could not get payment status")
			return nil, err
		}

		return pStatus.ToStatusAPI(), nil
	}
}

func (pb *paymentBusiness) InitiatePrompt(ctx context.Context, req *paymentV1.InitiatePromptRequest) (*commonv1.StatusResponse, error) {
	logger := pb.service.L(ctx).WithField("request", req)
	logger.Info("handling initiate prompt request")

	account := models.Account{
		AccountNumber: req.GetRecipientAccount().GetAccountNumber(),
		CountryCode:   req.GetRecipientAccount().GetCountryCode(),
		Name:          req.GetRecipientAccount().GetName(),
	}
	// Initialize Prompt model
	p := &models.Prompt{
		ID:                   req.GetId(),
		SourceID:             req.GetSource().GetProfileId(),
		SourceProfileType:    req.GetSource().GetProfileType(),
		SourceContactID:      req.GetSource().GetContactId(),
		RecipientID:          req.GetRecipient().GetProfileId(),
		RecipientProfileType: req.GetRecipient().GetProfileType(),
		RecipientContactID:   req.GetRecipient().GetContactId(),
		Amount:               &money.Money{CurrencyCode: req.GetAmount().GetCurrencyCode(), Units: req.GetAmount().GetUnits()},
		DateCreated:          time.Now().Format("2006-01-02 15:04:05"),
		DeviceID:             req.GetDeviceId(),
		State:                int32(commonv1.STATE_CREATED.Number()),
		Status:               int32(commonv1.STATUS_QUEUED.Number()),
		Account:              &account,
	}
	// Generate a unique transaction reference (6 chars - letter prefix + 5 digits)
	transactionRef := generateTransactionRef()
	if req.GetId() == "" {
		p.ID = transactionRef
	}
	p.Extra["transaction_ref"] = transactionRef

	// Send the STK/USSD push request to Jenga API asynchronously
	go func() {
		// Create new background context for async processing
		asyncCtx := context.Background()
		asyncLogger := pb.service.L(asyncCtx).WithField("function", "AsyncSTKPush").WithField("promptId", p.GetID())
		asyncLogger.Info("Starting async STK/USSD push request")

		// Format the current date and amount for the API
		currentDate := time.Now().Format("2006-01-02")
		amountStr := fmt.Sprintf("%.2f", float64(req.GetAmount().GetUnits())/100)
		callbackURL := os.Getenv("CALLBACK_URL")
		// Prepare the payload for Jenga API
		stkPayload := map[string]interface{}{
			"merchant": map[string]string{
				"accountNumber": account.AccountNumber,
				"countryCode":   account.CountryCode,
				"name":          account.Name,
			},
			"payment": map[string]string{
				"ref":          transactionRef,
				"amount":       amountStr,
				"currency":     req.GetAmount().GetCurrencyCode(),
				"telco":        req.Extra["telco"],
				"mobileNumber": req.GetSource().GetContactId(),
				"date":         currentDate,
				"callBackUrl":  callbackURL,
				"pushType":     req.Extra["pushType"],
			},
		}

		// Convert to JSON
		jsonData, err := json.Marshal(stkPayload)
		if err != nil {
			asyncLogger.WithError(err).Error("Failed to marshal STK request payload")
			return
		}

		// Get API endpoint from environment
		apiEndpoint := os.Getenv("JENGA_API_ENDPOINT")
		if apiEndpoint == "" {
			apiEndpoint = "https://uat.finserve.africa/v3-apis/transaction-api/v3.0/stk/push"
		}

		// Create the HTTP client with timeout
		client := &http.Client{Timeout: 30 * time.Second}

		// Create HTTP request
		httpReq, err := http.NewRequestWithContext(asyncCtx, "POST", apiEndpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			asyncLogger.WithError(err).Error("Failed to create HTTP request")
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")

		// Add authentication if configured
		authToken := os.Getenv("JENGA_API_TOKEN")
		if authToken != "" {
			httpReq.Header.Set("Authorization", "Bearer "+authToken)
		}

		// Send the request
		asyncLogger.WithField("payload", string(jsonData)).Info("Sending STK push request")
		resp, err := client.Do(httpReq)

		if err != nil {
			asyncLogger.WithError(err).Error("Failed to send STK push request")
			return
		}
		defer resp.Body.Close()

		// Read and process response
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			asyncLogger.WithError(err).Error("Failed to read response body")
			return
		}

		asyncLogger.WithFields(map[string]interface{}{
			"statusCode": resp.StatusCode,
			"body":       string(respBody),
			"reference":  transactionRef,
		}).Info("Received STK push response")
	}()

	//Emit events for Prompt
	event := events.PromptSave{}
	err := pb.service.Emit(ctx, event.Name(), p)
	if err != nil {
		logger.WithError(err).Warn("could not emit prompt save")
		return nil, err
	}

	// Set initial PromptStatus
	pStatus := models.PromptStatus{
		PromptID: p.GetID(),
		State:    int32(commonv1.STATE_CREATED.Number()),
		Status:   int32(commonv1.STATUS_QUEUED.Number()),
	}

	eventStatus := events.PromptStatusSave{}
	err = pb.service.Emit(ctx, eventStatus.Name(), pStatus)
	if err != nil {
		logger.WithError(err).Warn("could not emit prompt status save")
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}

// validateAmountAndCost validates the amount and cost fields of the Payment.
func (pb *paymentBusiness) validateAmountAndCost(message *paymentV1.Payment, p *models.Payment, c *models.Cost) {
	if message.GetAmount().Units <= 0 || message.GetAmount().CurrencyCode == "" {
		return
	}

	p.Amount = decimal.NullDecimal{
		Valid:   true,
		Decimal: decimal.NewFromFloat(float64(message.GetAmount().Units)),
	}
	p.Currency = message.GetAmount().CurrencyCode

	if message.GetCost().CurrencyCode == "" {
		return
	}

	c.Amount = decimal.NullDecimal{
		Valid:   true,
		Decimal: decimal.NewFromFloat(float64(message.GetCost().Units)),
	}
	c.Currency = message.GetCost().CurrencyCode
}

// generateTransactionRef creates a unique 6-character reference for Jenga API
func generateTransactionRef() string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	timeComponent := timestamp % 1000000
	asciiChar := 65 + ((timestamp / 1000000) % 26)
	return fmt.Sprintf("%c%05d", rune(asciiChar), timeComponent%100000)
}
