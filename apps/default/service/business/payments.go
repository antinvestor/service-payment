package business

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	partitionV1 "github.com/antinvestor/apis/go/partition/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments/service/events"
	"github.com/antinvestor/service-payments/service/models"
	"github.com/antinvestor/service-payments/service/repository"
	"github.com/antinvestor/service-payments/service/utility"
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

	c := &models.Cost{
		Amount: decimal.NullDecimal{
			Valid:   true,
			Decimal: utility.FromMoney(message.GetCost()),
		},
		Currency: message.GetCost().CurrencyCode,
	}
	c.GenID(ctx)

	if message.GetId() == "" {
		p.GenID(ctx)
	}

	pb.validateAmountAndCost(message, p, c)
	
	// Save cost separately and add its ID to payment
	costEvent := events.CostSave{Service: pb.service }
	if err := pb.service.Emit(ctx, costEvent.Name(), c); err != nil {
		pb.service.Log(ctx).WithError(err).Warn("could not emit cost event")
		return nil, err
	}
	p.CostIDs = []string{c.ID}

	event := events.PaymentSave{Service: pb.service}
	if err := pb.service.Emit(ctx, event.Name(), p); err != nil {
		pb.service.Log(ctx).WithError(err).Warn("could not emit payment event")
		return nil, err
	}

	// Unified status
	status := &models.Status{
		EntityID:   p.GetID(),
		EntityType: "payment",
		State:      int32(commonv1.STATE_CREATED.Number()),
		Status:     int32(commonv1.STATUS_QUEUED.Number()),
		Extra:      make(datatypes.JSONMap),
	}
	status.GenID(ctx)
	statusEvent := events.StatusSave{Service: pb.service}
	if err := pb.service.Emit(ctx, statusEvent.Name(), status); err != nil {
		pb.service.Log(ctx).WithError(err).Warn("could not emit status event")
		return nil, err
	}

	return &commonv1.StatusResponse{
		Id:     status.EntityID,
		State:  commonv1.STATE(status.State),
		Status: commonv1.STATUS(status.Status),
		Extras: frame.DBPropertiesToMap(status.Extra),
	}, nil
}

func (pb *paymentBusiness) Receive(ctx context.Context, message *paymentV1.Payment) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("request", message)
	logger.Info("handling receive request")

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
			Decimal: utility.FromMoney(message.GetCost()),
		},
		Currency: message.GetCost().CurrencyCode,
	}
	c.GenID(ctx)

	if message.GetId() == "" {
		p.GenID(ctx)
	}
	pb.validateAmountAndCost(message, p, c)
	
	// Save cost separately and add its ID to payment
	costEvent := events.CostSave{Service: pb.service }
	if err := pb.service.Emit(ctx, costEvent.Name(), c); err != nil {
		pb.service.Log(ctx).WithError(err).Warn("could not emit cost event")
		return nil, err
	}
	p.CostIDs = []string{c.ID}

	event := events.PaymentSave{Service: pb.service}
	if err := pb.service.Emit(ctx, event.Name(), p); err != nil {
		pb.service.Log(ctx).WithError(err).Warn("could not emit payment event")
		return nil, err
	}

	// Unified status
	status := &models.Status{
		EntityID:   p.GetID(),
		EntityType: "payment",
		State:      int32(commonv1.STATE_CREATED.Number()),
		Status:     int32(commonv1.STATUS_QUEUED.Number()),
		Extra:      make(datatypes.JSONMap),
	}
	status.GenID(ctx)
	statusEvent := events.StatusSave{Service: pb.service}
	if err := pb.service.Emit(ctx, statusEvent.Name(), status); err != nil {
		pb.service.Log(ctx).WithError(err).Warn("could not emit status event")
		return nil, err
	}

	return &commonv1.StatusResponse{
		Id:     status.EntityID,
		State:  commonv1.STATE(status.State),
		Status: commonv1.STATUS(status.Status),
		Extras: frame.DBPropertiesToMap(status.Extra),
	}, nil
}

func (pb *paymentBusiness) Status(ctx context.Context, statusReq *commonv1.StatusRequest) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("request", statusReq)
	logger.Info("handling status check request")

	statusRepo := repository.NewStatusRepository(ctx, pb.service)
	status, err := statusRepo.GetByEntity(ctx, statusReq.GetId(), statusReq.GetExtras()["entity_type"])
	if err != nil {
		logger.WithError(err).Error("could not get status")
		return nil, err
	}
	return &commonv1.StatusResponse{
		Id:     status.EntityID,
		State:  commonv1.STATE(status.State),
		Status: commonv1.STATUS(status.Status),
		Extras: frame.DBPropertiesToMap(status.Extra),
	}, nil
}

func (pb *paymentBusiness) StatusUpdate(ctx context.Context, req *commonv1.StatusUpdateRequest) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("request", req)
	logger.Info("handling unified status update request")

	entityType := req.GetExtras()["entity_type"]
	if entityType == "" {
		logger.Error("entity_type must be provided in extras for status update")
		return nil, fmt.Errorf("entity_type must be provided in extras for status update")
	}

	status := &models.Status{
		EntityID:   req.GetId(),
		EntityType: entityType,
		State:      int32(req.GetState()),
		Status:     int32(req.GetStatus()),
		Extra:      frame.DBPropertiesFromMap(req.GetExtras()),
	}
	status.GenID(ctx)

	statusEvent := events.StatusSave{Service: pb.service}
	if err := pb.service.Emit(ctx, statusEvent.Name(), status); err != nil {
		logger.WithError(err).Warn("could not emit status save")
		return nil, err
	}

	return &commonv1.StatusResponse{
		Id:     status.EntityID,
		State:  commonv1.STATE(status.State),
		Status: commonv1.STATUS(status.Status),
		Extras: frame.DBPropertiesToMap(status.Extra),
	}, nil
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
	paymentStatusRepo := repository.NewStatusRepository(ctx, pb.service)

	var responsesList []*paymentV1.Payment
	for _, p := range paymentList {
		var status *models.Status
		if p.ID != "" {
			status, err = paymentStatusRepo.GetByEntity(ctx, p.ID, "payment")
			if err != nil {
				logger.WithError(err).WithField("status_id", p.ID).Error("could not get status id")
				return err
			}
		}
		// Convert the payment model to the API response format
		result := p.ToApi(status, nil)
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

		event := events.PaymentSave{Service: pb.service}
		err = pb.service.Emit(ctx, event.Name(), p)
		if err != nil {
			logger.WithError(err).Warn("could not emit payment save")
			return nil, err
		}

		// Unified status
		status := &models.Status{
			EntityID:   p.GetID(),
			EntityType: "payment",
			State:      int32(commonv1.STATE_ACTIVE.Number()),
			Status:     int32(commonv1.STATUS_QUEUED.Number()),
			Extra:      make(datatypes.JSONMap),
		}
		status.GenID(ctx)
		statusEvent := events.StatusSave{Service: pb.service}
		err = pb.service.Emit(ctx, statusEvent.Name(), status)
		if err != nil {
			logger.WithError(err).Warn("could not emit status event")
			return nil, err
		}

		return &commonv1.StatusResponse{
			Id:     status.EntityID,
			State:  commonv1.STATE(status.State),
			Status: commonv1.STATUS(status.Status),
			Extras: frame.DBPropertiesToMap(status.Extra),
		}, nil
	} else {
		statusRepo := repository.NewStatusRepository(ctx, pb.service)
		status, err := statusRepo.GetByEntity(ctx, p.ID, "payment")
		if err != nil {
			logger.WithError(err).Warn("could not get payment status")
			return nil, err
		}
		return &commonv1.StatusResponse{
			Id:     status.EntityID,
			State:  commonv1.STATE(status.State),
			Status: commonv1.STATUS(status.Status),
			Extras: frame.DBPropertiesToMap(status.Extra),
		}, nil
	}
}

func (pb *paymentBusiness) InitiatePrompt(ctx context.Context, req *paymentV1.InitiatePromptRequest) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("request", req)
	logger.Info("handling initiate prompt request")

	// Build Account from request
	account := models.Account{
		AccountNumber: req.GetRecipientAccount().GetAccountNumber(),
		CountryCode:   req.GetRecipientAccount().GetCountryCode(),
		Name:          req.GetRecipientAccount().GetName(),
	}

	// Use AccountRepository to get or create the account
	var accountPtr *models.Account
	var err error
	accountPtr, err = repository.NewAccountRepository(ctx, pb.service).GetByAccountNumber(ctx, account.AccountNumber)
	if err != nil {
		// If not found, create the account
		account.GenID(ctx)
		event := events.AccountSave{Service: pb.service}
		err = pb.service.Emit(ctx, event.Name(), &account)
		if err != nil {
			logger.WithError(err).Warn("could not emit account save")
			return nil, err
		}
		accountPtr = &account
	}

	p := &models.Prompt{
		SourceID:             req.GetSource().GetProfileId(),
		SourceProfileType:    req.GetSource().GetProfileType(),
		SourceContactID:      req.GetSource().GetContactId(),
		RecipientID:          req.GetRecipient().GetProfileId(),
		RecipientProfileType: req.GetRecipient().GetProfileType(),
		RecipientContactID:   req.GetRecipient().GetContactId(),
		Amount:               decimal.NullDecimal{Valid: true, Decimal: utility.FromMoney(req.GetAmount())},
		DateCreated:          time.Now().Format("2006-01-02 15:04:05"),
		DeviceID:             req.GetDeviceId(),
		State:                int32(commonv1.STATE_CREATED.Number()),
		Status:               int32(commonv1.STATUS_QUEUED.Number()),
		AccountID:            accountPtr.ID,
		Account:              *accountPtr,
		Extra:                frame.DBPropertiesFromMap(req.Extra),
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
	p.Extra["mobile_number"] = req.GetSource().GetDetail()
	// Add telco and pushType information if provided

	event := events.PromptSave{Service: pb.service}
	err = pb.service.Emit(ctx, event.Name(), p)
	if err != nil {
		logger.WithError(err).Warn("could not emit prompt save")
		return nil, err
	}

	logger.WithField("promptId", p.ID).Info("Prompt saved and event emitted for STK/USSD processing")

	// Unified status usage
	status := &models.Status{
		EntityID:   p.ID,
		EntityType: "prompt",
		State:      int32(commonv1.STATE_CREATED.Number()),
		Status:     int32(commonv1.STATUS_QUEUED.Number()),
		Extra:      make(datatypes.JSONMap),
	}
	status.GenID(ctx)
	status.Extra["transaction_ref"] = transactionRef

	statusEvent := events.StatusSave{Service: pb.service}
	err = pb.service.Emit(ctx, statusEvent.Name(), status)
	if err != nil {
		logger.WithError(err).Warn("could not emit status save")
		return nil, err
	}


	err = pb.service.Publish(ctx, "initiate.prompt", p)
	if err != nil {
		logger.WithError(err).Warn("could not publish initiate-prompt")
		return nil, err
	}

	return &commonv1.StatusResponse{
		Id:     status.EntityID,
		State:  commonv1.STATE(status.State),
		Status: commonv1.STATUS(status.Status),
		Extras: frame.DBPropertiesToMap(status.Extra),
	}, nil
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
			profileName := c.GetSource().GetProfileName()
			firstName := profileName
			lastName := ""
			if len(profileName) > 0 {
				parts := strings.Fields(profileName)
				if len(parts) > 1 {
					firstName = parts[0]
					lastName = strings.Join(parts[1:], " ")
				} else {
					firstName = parts[0]
					lastName = ""
				}
			}

			customers = append(customers, models.Customer{
				FirstName:           firstName, // fallback: use ProfileName as FirstName
				LastName:            lastName,  // Not available in proto, unless split from ProfileName
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
		amount = utility.FromMoney(plReq.GetAmount())
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
	event := events.PaymentLinkSave{Service: pb.service}
	if err := pb.service.Emit(ctx, event.Name(), paymentLink); err != nil {
		logger.WithError(err).Warn("could not emit payment link save event")
		return nil, err
	}

	// Create PaymentLinkStatus
	status := &models.Status{
		EntityID:   paymentLink.ID,
		EntityType: "payment_link",
		State:      int32(commonv1.STATE_CREATED.Number()),
		Status:     int32(commonv1.STATUS_QUEUED.Number()),
		Extra:      make(map[string]interface{}),
	}
	status.GenID(ctx)
	statusEvent := events.StatusSave{Service: pb.service}
	if err := pb.service.Emit(ctx, statusEvent.Name(), status); err != nil {
		logger.WithError(err).Warn("could not emit payment link status event")
		return nil, err
	}

	err = pb.service.Publish(ctx, "create.payment.link", paymentLink)
	if err != nil {
		logger.WithError(err).Warn("could not publish create-payment-link")
		// Emit the status event even if publish fails
		status.State = int32(commonv1.STATE_INACTIVE.Number())
		status.Status = int32(commonv1.STATUS_FAILED.Number())
		status.Extra["error"] = err.Error()
		if err := pb.service.Emit(ctx, statusEvent.Name(), status); err != nil {
			logger.WithError(err).Warn("could not emit payment link status event after publish failure")
		}
		return nil, err
	}
	return &commonv1.StatusResponse{
		Id:     status.EntityID,
		State:  commonv1.STATE(status.State),
		Status: commonv1.STATUS(status.Status),
		Extras: frame.DBPropertiesToMap(status.Extra),
	}, nil
}

// validateAmountAndCost validates the amount and cost fields of the Payment.
func (pb *paymentBusiness) validateAmountAndCost(message *paymentV1.Payment, p *models.Payment, c *models.Cost) {
	if message.GetAmount().Units <= 0 || message.GetAmount().CurrencyCode == "" {
		return
	}

	p.Amount = decimal.NullDecimal{
		Valid:   true,
		Decimal: utility.FromMoney(message.GetAmount()),
	}
	p.Currency = message.GetAmount().CurrencyCode

	if message.GetCost().CurrencyCode == "" {
		return
	}

	c.Amount = decimal.NullDecimal{
		Valid:   true,
		Decimal: utility.FromMoney(message.GetCost()),
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
