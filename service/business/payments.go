package business

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	partitionV1 "github.com/antinvestor/apis/go/partition/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments/service/events"
	"github.com/antinvestor/service-payments/service/models"
	"github.com/antinvestor/service-payments/service/repository"
	"github.com/pitabwire/frame"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

type PaymentBusiness interface {
	Send(ctx context.Context, payment *paymentV1.Payment) (*commonv1.StatusResponse, error)
	Receive(ctx context.Context, payment *paymentV1.Payment) (*commonv1.StatusResponse, error)
	Status(ctx context.Context, status *commonv1.StatusRequest) (*commonv1.StatusResponse, error)
	StatusUpdate(ctx context.Context, req *commonv1.StatusUpdateRequest) (*commonv1.StatusResponse, error)
	Release(ctx context.Context, status *paymentV1.ReleaseRequest) (*commonv1.StatusResponse, error)
	Search(search *commonv1.SearchRequest, stream paymentV1.PaymentService_SearchServer) error
	InitiatePrompt(ctx context.Context, req *paymentV1.InitiatePromptRequest) (*commonv1.StatusResponse, error)
	CreatePaymentLink(ctx context.Context, req *paymentV1.CreatePaymentLinkRequest) (*commonv1.StatusResponse, error)
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
		pb.service.Log(ctx).WithError(err).Warn("could not emit payment event")
		return nil, err
	}

	eventStatus := events.PaymentStatusSave{}
	if err := pb.service.Emit(ctx, eventStatus.Name(), pStatus); err != nil {
		pb.service.Log(ctx).WithError(err).Warn("could not emit payment status event")
		return nil, err
	}
	return pStatus.ToStatusAPI(), nil
}

func (pb *paymentBusiness) Receive(ctx context.Context, message *paymentV1.Payment) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("request", message)
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
	// IMPORTANT: Use the payment's generated ID instead of the message.GetId() which might be empty
	pStatus := models.PaymentStatus{
		PaymentID: p.GetID(), // Use the payment's ID that was just generated/validated
		State:     int32(commonv1.STATE_CREATED.Number()),
		Status:    int32(commonv1.STATUS_QUEUED.Number()),
	}

	// Generate ID for the PaymentStatus
	pStatus.GenID(ctx)

	event := events.PaymentSave{}
	if err := pb.service.Emit(ctx, event.Name(), p); err != nil {
		pb.service.Log(ctx).WithError(err).Warn("could not emit payment event")
		return nil, err
	}

	eventStatus := events.PaymentStatusSave{}
	if err := pb.service.Emit(ctx, eventStatus.Name(), pStatus); err != nil {
		pb.service.Log(ctx).WithError(err).Warn("could not emit payment status event")
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}

func (pb *paymentBusiness) Status(ctx context.Context, status *commonv1.StatusRequest) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("request", status)
	logger.Info("handling status check request")

	// Define a slice of status handlers that will be tried in order
	statusHandlers := []func(context.Context, string) (*commonv1.StatusResponse, error){
		pb.getPaymentStatus,
		pb.getPromptStatus,
		pb.getPaymentLinkStatus,
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
	logger := pb.service.Log(ctx).WithField("request", req)
	logger.Info("handling status update request")

	// Check if the request has an explicit update_type specified
	updateType, hasUpdateType := req.GetExtras()["update_type"]
	if hasUpdateType {
		logger.WithField("update_type", updateType).Info("Request has explicit update type")

		// Handle explicit prompt update first
		switch updateType {
		case "prompt":
			logger.Info("Processing explicit prompt status update")
			statusResponse, err := pb.updatePromptStatus(ctx, req)
			if err == nil {
				return statusResponse, nil
			}
			logger.WithError(err).Error("Failed to update prompt status despite explicit type")
			// Return the error since this was explicitly a prompt update
			return nil, err
		case "payment":
			logger.Info("Processing explicit payment status update")
			statusResponse, err := pb.updatePaymentStatus(ctx, req)
			if err == nil {
				return statusResponse, nil
			}
			logger.WithError(err).Error("Failed to update payment status despite explicit type")
			// Return the error since this was explicitly a payment update
			return nil, err
		case "payment_link":
			logger.Info("Processing explicit payment link status update")
			statusResponse, err := pb.updatePaymentLinkStatus(ctx, req)
			if err == nil {
				return statusResponse, nil
			}
			logger.WithError(err).Error("Failed to update payment link status despite explicit type")
			// Return the error since this was explicitly a payment link update
			return nil, err
		}
	}

	// If no specific type was provided, try the standard order of handlers
	// Define a slice of status update handlers that will be tried in order
	statusUpdateHandlers := []func(context.Context, *commonv1.StatusUpdateRequest) (*commonv1.StatusResponse, error){
		pb.updatePaymentStatus,
		pb.updatePromptStatus,
		pb.updatePaymentLinkStatus,
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

// getPaymentStatus tries to get the status of a payment with the given ID.
func (pb *paymentBusiness) getPaymentStatus(ctx context.Context, id string) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("paymentId", id)

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

// getPaymentLinkStatus tries to get the status of a payment link with the given ID.

func (pb *paymentBusiness) getPaymentLinkStatus(ctx context.Context, id string) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("paymentLinkId", id)
	// Try to find a payment link with this ID
	paymentLinkRepo := repository.NewPaymentLinkRepository(ctx, pb.service)
	paymentLink, err := paymentLinkRepo.GetByID(ctx, id)
	if err != nil {
		logger.WithError(err).Error("could not get payment link by ID")
		return nil, err
	}
	// Get the payment link status
	paymentLinkStatusRepo := repository.NewPaymentLinkStatusRepository(ctx, pb.service)
	pStatus, err := paymentLinkStatusRepo.GetByID(ctx, paymentLink.ID)
	if err != nil {
		logger.WithError(err).Error("could not get payment link status")
		return nil, err
	}
	return pStatus.ToStatusAPI(), nil
}

// getPromptStatus tries to get the status of a prompt with the given ID.
func (pb *paymentBusiness) getPromptStatus(ctx context.Context, id string) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("promptId", id)

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

// updatePaymentStatus tries to update the status of a payment with the given ID.
func (pb *paymentBusiness) updatePaymentStatus(ctx context.Context, req *commonv1.StatusUpdateRequest) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("paymentId", req.GetId())

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

// updatePromptStatus tries to update the status of a prompt with the given ID.
func (pb *paymentBusiness) updatePromptStatus(ctx context.Context, req *commonv1.StatusUpdateRequest) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("promptId", req.GetId())

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

// update payment link status.
func (pb *paymentBusiness) updatePaymentLinkStatus(ctx context.Context, req *commonv1.StatusUpdateRequest) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("paymentLinkId", req.GetId())

	// Try to find a payment link with this ID
	paymentLinkRepo := repository.NewPaymentLinkRepository(ctx, pb.service)
	paymentLink, err := paymentLinkRepo.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	// Create a payment link status update
	pStatus := models.PaymentLinkStatus{
		PaymentLinkID: paymentLink.ID,
		State:         int32(req.GetState()),
		Status:        int32(req.GetStatus()),
		Extra:         frame.DBPropertiesFromMap(req.GetExtras()),
	}
	pStatus.GenID(ctx)

	// Emit the payment link status event
	eventStatus := events.PaymentLinkStatusSave{}
	err = pb.service.Emit(ctx, eventStatus.Name(), pStatus)
	if err != nil {
		logger.WithError(err).Error("could not emit payment link status event")
		return nil, err
	}
	return pStatus.ToStatusAPI(), nil
}

func (pb *paymentBusiness) Search(search *commonv1.SearchRequest,
	stream paymentV1.PaymentService_SearchServer) error {
	logger := pb.service.Log(stream.Context()).WithField("request", search)
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
	logger := pb.service.Log(ctx).WithField("request", paymentReq)
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
	logger := pb.service.Log(ctx).WithField("request", req)
	logger.Info("handling initiate prompt request")

	account := models.Account{
		AccountNumber: req.GetRecipientAccount().GetAccountNumber(),
		CountryCode:   req.GetRecipientAccount().GetCountryCode(),
		Name:          req.GetRecipientAccount().GetName(),
	}
	// Initialize Prompt model
	p := &models.Prompt{
		SourceID:             req.GetSource().GetProfileId(),
		SourceProfileType:    req.GetSource().GetProfileType(),
		SourceContactID:      req.GetSource().GetContactId(),
		RecipientID:          req.GetRecipient().GetProfileId(),
		RecipientProfileType: req.GetRecipient().GetProfileType(),
		RecipientContactID:   req.GetRecipient().GetContactId(),
		Amount:               decimal.NullDecimal{Valid: true, Decimal: decimal.NewFromFloat(float64(req.GetAmount().GetUnits()))},
		DateCreated:          time.Now().Format("2006-01-02 15:04:05"),
		DeviceID:             req.GetDeviceId(),
		State:                int32(commonv1.STATE_CREATED.Number()),
		Status:               int32(commonv1.STATUS_QUEUED.Number()),
		Account: func() datatypes.JSON {
			accountJSON, err := json.Marshal(account)
			if err != nil {
				logger.WithError(err).Error("failed to marshal account to JSON")
				return datatypes.JSON("{}")
			}
			return accountJSON
		}(),
		Extra: make(datatypes.JSONMap),
	}

	// Generate a unique transaction reference (6 chars - letter prefix + 5 digits)
	transactionRef := generateTransactionRef()

	// First explicitly set the provided ID if one was given
	if req.GetId() != "" {
		p.ID = req.GetId()
	}

	if p.ID == "" {
		p.GenID(ctx)
		p.ID = p.GetID()
	}

	logger.WithField("promptId", p.ID).Info("Prompt ID set")

	p.Extra["transaction_ref"] = transactionRef
	p.Extra["currency"] = req.GetAmount().GetCurrencyCode()

	// Add telco and pushType information if provided
	if telco, ok := req.Extra["telco"]; ok {
		p.Extra["telco"] = telco
	}

	if pushType, ok := req.Extra["pushType"]; ok {
		p.Extra["pushType"] = pushType
	}

	p.Extra["mobile_number"] = req.GetSource().GetContactId()

	event := events.PromptSave{}
	err := pb.service.Emit(ctx, event.Name(), p)
	if err != nil {
		logger.WithError(err).Warn("could not emit prompt save")
		return nil, err
	}

	logger.WithField("promptId", p.ID).Info("Prompt saved and event emitted for STK/USSD processing")

	pStatus := models.PromptStatus{
		PromptID: p.ID,
		State:    int32(commonv1.STATE_CREATED.Number()),
		Status:   int32(commonv1.STATUS_QUEUED.Number()),
		Extra:    make(datatypes.JSONMap),
	}
	pStatus.ID = p.GetID()
	pStatus.Extra["transaction_ref"] = transactionRef

	eventStatus := events.PromptStatusSave{}
	err = pb.service.Emit(ctx, eventStatus.Name(), pStatus)
	if err != nil {
		logger.WithError(err).Warn("could not emit prompt status save")
		return nil, err
	}

	err = pb.service.Publish(ctx, "initiate.prompt", p)
	if err != nil {
		logger.WithError(err).Warn("could not publish initiate-prompt")
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}

func (pb *paymentBusiness) CreatePaymentLink(ctx context.Context, req *paymentV1.CreatePaymentLinkRequest) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("request", req)
	logger.Info("handling create payment link request")

	// Validate required fields
	if req == nil || req.GetPaymentLink() == nil {
		logger.Error("missing payment link payload")
		return nil, fmt.Errorf("missing payment link payload")
	}

	plReq := req.GetPaymentLink()

	// Marshal customers to JSON
	var customersJSON datatypes.JSON
	if len(req.GetCustomers()) > 0 {
		customers := make([]models.Customer, 0, len(req.GetCustomers()))
		for _, c := range req.GetCustomers() {
			customers = append(customers, models.Customer{
				FirstName:           c.GetSource().GetProfileName(), // fallback: use ProfileName as FirstName
				LastName:            "",                             // Not available in proto, unless split from ProfileName
				Email:               c.GetSource().GetExtras()["email"],
				PhoneNumber:         c.GetSource().GetContactId(),
				FirstAddress:        c.GetFirstAddress(),
				CountryCode:         c.GetCountryCode(),
				PostalOrZipCode:     c.GetPostalOrZipCode(),
				CustomerExternalRef: c.GetCustomerExternalRef(),
			})
		}
		b, err := json.Marshal(customers)
		if err != nil {
			logger.WithError(err).Error("failed to marshal customers")
			return nil, err
		}
		customersJSON = b
	}

	// Marshal notifications to JSON
	var notificationsJSON datatypes.JSON
	if len(req.GetNotifications()) > 0 {
		notificationTypes := make([]models.NotificationType, 0, len(req.GetNotifications()))
		for _, n := range req.GetNotifications() {
			switch n {
			case paymentV1.NotificationType_NOTIFICATION_TYPE_EMAIL:
				notificationTypes = append(notificationTypes, models.NotificationTypeEmail)
			case paymentV1.NotificationType_NOTIFICATION_TYPE_SMS:
				notificationTypes = append(notificationTypes, models.NotificationTypeSMS)
			}
		}
		b, err := json.Marshal(notificationTypes)
		if err != nil {
			logger.WithError(err).Error("failed to marshal notifications")
			return nil, err
		}
		notificationsJSON = b
	}

	// Parse dates
	expiryDate, err := time.Parse("2006-01-02", plReq.GetExpiryDate())
	if err != nil {
		expiryDate = time.Now().Add(1 * 24 * time.Hour) // default: 1 days from now
	}
	saleDate, err := time.Parse("2006-01-02", plReq.GetSaleDate())
	if err != nil {
		saleDate = time.Now()
	}

	// Parse amount
	amount := decimal.NewFromInt(0)
	if plReq.GetAmount() != nil {
		amount = decimal.NewFromInt(plReq.GetAmount().GetUnits())
	}

	// Build PaymentLink model
	paymentLink := &models.PaymentLink{
		ExpiryDate:      expiryDate,
		SaleDate:        saleDate,
		PaymentLinkType: plReq.GetPaymentLinkType(),
		SaleType:        plReq.GetSaleType(),
		Name:            plReq.GetName(),
		Description:     plReq.GetDescription(),
		ExternalRef:     plReq.GetExternalRef(),
		PaymentLinkRef:  plReq.GetPaymentLinkRef(),
		RedirectURL:     plReq.GetRedirectUrl(),
		AmountOption:    plReq.GetAmountOption(),
		Amount:          amount,
		Currency:        plReq.GetCurrency(),
		Customers:       customersJSON,
		Notifications:   notificationsJSON,
	}

	// Set ID if provided
	if plReq.GetId() != "" {
		paymentLink.ID = plReq.GetId()
	}

	// Generate ID if not set
	if paymentLink.ID == "" {
		paymentLink.GenID(ctx)
	}

	// Save PaymentLink (emit event)
	event := events.PaymentLinkSave{}
	if err := pb.service.Emit(ctx, event.Name(), paymentLink); err != nil {
		logger.WithError(err).Warn("could not emit payment link save event")
		return nil, err
	}

	// Create PaymentLinkStatus
	pStatus := models.PaymentLinkStatus{
		PaymentLinkID: paymentLink.ID,
		State:         int32(commonv1.STATE_CREATED.Number()),
		Status:        int32(commonv1.STATUS_QUEUED.Number()),
		Extra:         make(map[string]interface{}),
	}
	pStatus.GenID(ctx)

	// Emit PaymentLinkStatus event
	eventStatus := events.PaymentLinkStatusSave{}
	if err := pb.service.Emit(ctx, eventStatus.Name(), pStatus); err != nil {
		logger.WithError(err).Warn("could not emit payment link status event")
		return nil, err
	}

	err = pb.service.Publish(ctx, "create.payment.link", paymentLink)
	if err != nil {
		logger.WithError(err).Warn("could not publish create-payment-link")
		// Emit the status event even if publish fails
		eventStatus = events.PaymentLinkStatusSave{}
		pStatus.State = int32(commonv1.STATE_INACTIVE.Number())
		pStatus.Status = int32(commonv1.STATUS_FAILED.Number())
		pStatus.Extra["error"] = err.Error()
		if err := pb.service.Emit(ctx, eventStatus.Name(), pStatus); err != nil {
			logger.WithError(err).Warn("could not emit payment link status event after publish failure")
		}
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

// generateTransactionRef creates a unique 6-character reference for Jenga API.
func generateTransactionRef() string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	timeComponent := timestamp % 1000000
	asciiChar := 65 + ((timestamp / 1000000) % 26)
	return fmt.Sprintf("%c%05d", rune(asciiChar), timeComponent%100000)
}
