package business

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/ledger/connectrpc/go/ledger/v1/ledgerv1connect"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/ledger/v1"
	"buf.build/gen/go/antinvestor/partition/connectrpc/go/partition/v1/partitionv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/payment/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/profile/v1/profilev1connect"
	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/apis/go/ledger"
	"github.com/antinvestor/apis/go/partition"
	"github.com/antinvestor/apis/go/profile"
	"github.com/antinvestor/service-payments/service/events"
	"github.com/antinvestor/service-payments/service/models"
	"github.com/antinvestor/service-payments/service/repository"
	"github.com/antinvestor/service-payments/service/utility"
	"github.com/pitabwire/frame"
	"github.com/pitabwire/frame/data"
	fevents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/security"
	"github.com/pitabwire/frame/workerpool"
	"github.com/pitabwire/util"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

func NewPaymentBusiness(
	ctx context.Context,
	service *frame.Service,
	profileCli profilev1connect.ProfileServiceClient,
	partitionCli partitionv1connect.PartitionServiceClient,
	ledgerCli ledgerv1connect.LedgerServiceClient,
	paymentRepo repository.PaymentRepository,
	statusRepo repository.StatusRepository,
	costRepo repository.CostRepository,
	accountRepo repository.AccountRepository,
	promptRepo repository.PromptRepository,
	paymentLinkRepo repository.PaymentLinkRepository,
) (PaymentBusiness, error) {
	if service == nil {
		return nil, ErrInitializationFail
	}

	workMan := service.WorkManager()

	return &paymentBusiness{
		service:         service,
		profileCli:      profileCli,
		partitionCli:    partitionCli,
		ledgerCli:       ledgerCli,
		paymentRepo:     paymentRepo,
		statusRepo:      statusRepo,
		costRepo:        costRepo,
		accountRepo:     accountRepo,
		promptRepo:      promptRepo,
		paymentLinkRepo: paymentLinkRepo,
		workMan:         workMan,
	}, nil
}

type paymentBusiness struct {
	service         *frame.Service
	eventMan        fevents.Manager
	profileCli      profilev1connect.ProfileServiceClient
	partitionCli    partitionv1connect.PartitionServiceClient
	ledgerCli       ledgerv1connect.LedgerServiceClient
	paymentRepo     repository.PaymentRepository
	statusRepo      repository.StatusRepository
	costRepo        repository.CostRepository
	accountRepo     repository.AccountRepository
	promptRepo      repository.PromptRepository
	paymentLinkRepo repository.PaymentLinkRepository
	workMan         workerpool.Manager
}

func (pb *paymentBusiness) Send(ctx context.Context, message *paymentv1.Payment) (*commonv1.StatusResponse, error) {
	p := &models.Payment{
		SenderProfileType:    message.GetSource().GetProfileType(),
		SenderProfileID:      message.GetSource().GetProfileId(),
		SenderContactID:      message.GetSource().GetContactId(),
		RecipientProfileType: message.GetRecipient().GetProfileType(),
		RecipientProfileID:   message.GetRecipient().GetProfileId(),
		RecipientContactID:   message.GetRecipient().GetContactId(),
		ReferenceID:          message.GetReferenceId(),
		BatchID:              message.GetBatchId(),
		RouteID:              message.GetRoute(),
		PaymentType:          "Bank Transfers",
		OutBound:             true,
	}

	c := &models.Cost{
		Amount: decimal.NullDecimal{
			Valid:   true,
			Decimal: utility.FromMoney(message.GetCost()),
		},
		Currency: message.GetCost().GetCurrencyCode(),
	}
	c.GenID(ctx)

	if message.GetId() == "" {
		p.GenID(ctx)
	}

	pb.validateAmountAndCost(message, p, c)

	// Save cost separately and add its ID to payment
	costEvent := events.CostSave{Service: pb.service}
	if err := pb.eventMan.Emit(ctx, costEvent.Name(), c); err != nil {
		pb.service.Log(ctx).WithError(err).Warn("could not emit cost event")
		return nil, err
	}
	p.CostIDs = []string{c.ID}

	event := events.PaymentSave{Service: pb.service}
	if err := pb.eventMan.Emit(ctx, event.Name(), p); err != nil {
		pb.service.Log(ctx).WithError(err).Warn("could not emit payment event")
		return nil, err
	}

	senderTel := ""
	if message.GetSource() != nil {
		senderTel = message.GetSource().GetDetail()
	}

	// try member name from source profile name or extras
	memberName := ""
	if message.GetSource() != nil {
		memberName = message.GetSource().GetProfileName()
		if memberName == "" {
			if v, ok := message.GetSource().GetExtras()["member_name"]; ok {
				memberName = v
			}
		}
	}

	// try group name from payment-level extras first, then source extras
	groupName := ""
	if v, ok := message.GetSource().GetExtras()["group_name"]; ok && v != "" {
		groupName = v
	} else if message.GetSource() != nil {
		if v, ok := message.GetSource().GetExtras()["group_name"]; ok {
			groupName = v
		}
	}

	// Create ledger transaction for outbound payment
	if pb.ledgerCli != nil && p.Amount.Valid {
		if err := pb.createDepositStep1(ctx, p, senderTel, groupName, memberName); err != nil {
			pb.service.Log(ctx).WithError(err).Warn("could not create ledger transaction for send")
			// Don't fail the payment if ledger fails, just log the error
		}
	}

	// Create status using repository
	status := &models.Status{
		EntityID:   p.GetID(),
		EntityType: "payment",
		State:      int32(commonv1.STATE_CREATED.Number()),
		Status:     int32(commonv1.STATUS_QUEUED.Number()),
		Extra:      make(data.JSONMap),
	}
	status.GenID(ctx)

	err = pb.statusRepo.Create(ctx, status)
	if err != nil {
		pb.service.Log(ctx).WithError(err).Warn("could not save status")
		return nil, err
	}

	return status.ToAPI(), nil
}

func (pb *paymentBusiness) Receive(ctx context.Context, message *paymentv1.Payment) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("request", message)
	logger.Info("handling receive request")

	p := &models.Payment{
		SenderProfileType:    message.GetSource().GetProfileType(),
		SenderProfileID:      message.GetSource().GetProfileId(),
		SenderContactID:      message.GetSource().GetContactId(),
		RecipientProfileType: message.GetRecipient().GetProfileType(),
		RecipientProfileID:   message.GetRecipient().GetProfileId(),
		RecipientContactID:   message.GetRecipient().GetContactId(),
		ReferenceID:          message.GetReferenceId(),
		BatchID:              message.GetBatchId(),
		RouteID:              message.GetRoute(),
		OutBound:             false,
	}

	c := &models.Cost{
		Amount: decimal.NullDecimal{
			Valid:   true,
			Decimal: utility.FromMoney(message.GetCost()),
		},
		Currency: message.GetCost().GetCurrencyCode(),
	}
	c.GenID(ctx)

	if message.GetId() == "" {
		p.GenID(ctx)
	}
	pb.validateAmountAndCost(message, p, c)

	// Save cost separately and add its ID to payment
	costEvent := events.CostSave{Service: pb.service}
	if err := pb.eventMan.Emit(ctx, costEvent.Name(), c); err != nil {
		pb.service.Log(ctx).WithError(err).Warn("could not emit cost event")
		return nil, err
	}
	p.CostIDs = []string{c.ID}

	event := events.PaymentSave{Service: pb.service}
	if err := pb.eventMan.Emit(ctx, event.Name(), p); err != nil {
		pb.service.Log(ctx).WithError(err).Warn("could not emit payment event")
		return nil, err
	}

	senderTel := ""
	if message.GetSource() != nil {
		senderTel = message.GetSource().GetDetail()
	}

	// try member name from source profile name or extras
	memberName := ""
	if message.GetSource() != nil {
		memberName = message.GetSource().GetProfileName()
		if memberName == "" {
			if v, ok := message.GetSource().GetExtras()["member_name"]; ok {
				memberName = v
			}
		}
	}

	// try group name from payment-level extras first, then source extras
	groupName := ""
	if v, ok := message.GetSource().GetExtras()["group_name"]; ok && v != "" {
		groupName = v
	} else if message.GetSource() != nil {
		if v, ok := message.GetSource().GetExtras()["group_name"]; ok {
			groupName = v
		}
	}

	// Create ledger transaction for inbound payment
	if pb.ledgerCli != nil && p.Amount.Valid {
		if err := pb.createDepositStep1(ctx, p, senderTel, groupName, memberName); err != nil {
			pb.service.Log(ctx).WithError(err).Warn("could not create ledger transaction for receive")
			// Don't fail the payment if ledger fails, just log the error
		}
	}

	// Unified status
	status := &models.Status{
		EntityID:   p.GetID(),
		EntityType: "payment",
		State:      int32(commonv1.STATE_CREATED.Number()),
		Status:     int32(commonv1.STATUS_QUEUED.Number()),
		Extra:      make(data.JSONMap),
	}
	status.GenID(ctx)

	if err := pb.eventMan.Emit(ctx, events.EventNameStatusSave, status); err != nil {
		pb.service.Log(ctx).WithError(err).Warn("could not emit status event")
		return nil, err
	}

	return status.ToAPI(), nil
}

func (pb *paymentBusiness) Status(
	ctx context.Context,
	statusReq *commonv1.StatusRequest,
) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("request", statusReq)
	logger.Info("handling status check request")

	status, err := pb.statusRepo.GetByEntity(ctx, statusReq.GetId(), statusReq.GetExtras()["entity_type"])
	if err != nil {
		logger.WithError(err).Error("could not get status")
		return nil, err
	}
	return status.ToAPI(), nil
}

func (pb *paymentBusiness) StatusUpdate(
	ctx context.Context,
	req *commonv1.StatusUpdateRequest,
) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("request", req)
	logger.Info("handling unified status update request")

	var extras data.JSONMap
	extras = extras.FromProtoStruct(req.GetExtras())
	
	entityType := extras.GetString("entity_type")
	if entityType == "" {
		logger.Error("entity_type must be provided in extras for status update")
		return nil, errors.New("entity_type must be provided in extras for status update")
	}

	status := &models.Status{
		EntityID:   req.GetId(),
		EntityType: entityType,
		State:      int32(req.GetState()),
		Status:     int32(req.GetStatus()),
		Extra:      extras,
	}
	status.GenID(ctx)


	if err := pb.eventMan.Emit(ctx, events.EventNameStatusSave, status); err != nil {
		logger.WithError(err).Warn("could not emit status save")
		return nil, err
	}

	return status.ToAPI(), nil
}

func (pb *paymentBusiness) Search(ctx context.Context, searchQuery *commonv1.SearchRequest) (workerpool.JobResultPipe[[]*paymentv1.Payment], error) {
	logger := util.Log(ctx).WithField("request", search)
	logger.Debug("handling payment search request")

	// Extract the context and JWT token
	jwtToken := security.JwtFromContext(ctx)
	logger.WithField("jwt", jwtToken).Debug("auth jwt supplied")

	var paymentList []*models.Payment
	var err error

	// Handle search by ID or by general query
	if search.GetIdQuery() != "" {
		// Search by ID
		payment, err0 := pb.paymentRepo.GetByID(ctx, search.GetIdQuery())
		if err0 != nil {
			return err0
		}

		paymentList = append(paymentList, payment)
	} else {
		// General search query - convert to new search format
		query := data.NewSearchQuery(
			data.WithSearchQuery(search.GetQuery()),
			data.WithSearchLimit(100), // default limit
		)

		// For now, use legacy search method in repository
		paymentList, err = pb.paymentRepo.Search(ctx, query)
		if err != nil {
			logger.WithError(err).Error("failed to search payments")
			return nil, err
		}
	}

	var responsesList []*paymentv1.Payment
	for _, p := range paymentList {
		var status *models.Status
		if p.ID != "" {
			status, err = pb.statusRepo.GetByEntity(ctx, p.ID, "payment")
			if err != nil {
				logger.WithError(err).WithField("status_id", p.ID).Error("could not get status id")
				return nil, err
			}
		}
		// Convert the payment model to the API response format
		result := p.ToAPI(status, nil)
		responsesList = append(responsesList, result)
	}

	// Send the search response back to the client
	err = stream.Send(&paymentv1.SearchResponse{Data: responsesList})
	if err != nil {
		logger.WithError(err).Warn("unable to send a result")
	}

	return nil
}

func (pb *paymentBusiness) Release(
	ctx context.Context,
	paymentReq *paymentv1.ReleaseRequest,
) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("request", paymentReq)
	logger.Debug("handling release request")

	p, err := pb.paymentRepo.GetByID(ctx, paymentReq.GetId())
	if err != nil {
		logger.WithError(err).Warn("could not fetch payment by id")
		return nil, err
	}

	if !p.IsReleased() {
		releaseDate := time.Now()
		p.ReleasedAt = &releaseDate

		event := events.PaymentSave{Service: pb.service}
		err = pb.eventMan.Emit(ctx, event.Name(), p)
		if err != nil {
			logger.WithError(err).Warn("could not update payment")
			return nil, err
		}

		// Create status using repository
		status := &models.Status{
			EntityID:   p.GetID(),
			EntityType: "payment",
			State:      int32(commonv1.STATE_ACTIVE.Number()),
			Status:     int32(commonv1.STATUS_QUEUED.Number()),
			Extra:      make(data.JSONMap),
		}
		status.GenID(ctx)

		err = pb.eventMan.Emit(ctx, events.EventNameStatusSave, status)
		if err != nil {
			logger.WithError(err).Warn("could not save status")
			return nil, err
		}

		return &commonv1.StatusResponse{
			Id:     status.EntityID,
			State:  commonv1.STATE(status.State),
			Status: commonv1.STATUS(status.Status),
			Extras: status.Extra),
		}, nil
	} else {
		status, err := pb.statusRepo.GetByEntity(ctx, p.ID, "payment")
		if err != nil {
			logger.WithError(err).Warn("could not get payment status")
			return nil, err
		}
		return &commonv1.StatusResponse{
			Id:     status.EntityID,
			State:  commonv1.STATE(status.State),
			Status: commonv1.STATUS(status.Status),
			Extras: status.Extra),
		}, nil
	}
}

func (pb *paymentBusiness) InitiatePrompt(
	ctx context.Context,
	req *paymentv1.InitiatePromptRequest,
) (*commonv1.StatusResponse, error) {
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
	accountPtr, err = pb.accountRepo.GetByAccountNumber(ctx, account.AccountNumber)
	if err != nil {
		// If not found, create the account
		account.GenID(ctx)
		err = pb.eventMan.Emit(ctx, events.EventNameAccountSave, &account)
		if err != nil {
			logger.WithError(err).Warn("could not create account")
			return nil, err
		}
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
		Extra:                req.GetExtra().AsMap(),
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
	err = pb.eventMan.Emit(ctx, event.Name(), p)
	if err != nil {
		logger.WithError(err).Warn("could not emit prompt save")
		return nil, err
	}

	logger.WithField("promptId", p.ID).Info("Prompt saved and event emitted for STK/USSD processing")

	// Create status using repository
	status := &models.Status{
		EntityID:   p.ID,
		EntityType: "prompt",
		State:      int32(commonv1.STATE_CREATED.Number()),
		Status:     int32(commonv1.STATUS_QUEUED.Number()),
		Extra:      make(data.JSONMap),
	}
	status.GenID(ctx)
	status.Extra["transaction_ref"] = transactionRef


	err = pb.eventMan.Emit(ctx, events.EventNameStatusSave, status)
	if err != nil {
		logger.WithError(err).Warn("could not save status")
		return nil, err
	}

	err = pb.service.Publish(ctx, "initiate.prompt", p)
	if err != nil {
		logger.WithError(err).Warn("could not publish initiate-prompt")
		return nil, err
	}

	return status.ToAPI(), nil
}

func (pb *paymentBusiness) CreatePaymentLink(
	ctx context.Context,
	req *paymentv1.CreatePaymentLinkRequest,
) (*commonv1.StatusResponse, error) {
	logger := pb.service.Log(ctx).WithField("request", req)
	logger.Info("handling create payment link request")

	// Validate required fields
	if req == nil || req.GetPaymentLink() == nil {
		logger.Error("missing payment link payload")
		return nil, errors.New("missing payment link payload")
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
			case paymentv1.NotificationType_NOTIFICATION_TYPE_EMAIL:
				notificationTypes = append(notificationTypes, models.NotificationTypeEmail)
			case paymentv1.NotificationType_NOTIFICATION_TYPE_SMS:
				notificationTypes = append(notificationTypes, models.NotificationTypeSMS)
			case paymentv1.NotificationType_NOTIFICATION_TYPE_UNSPECIFIED:
				// Skip unspecified notification types
				continue
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
	if emitErr := pb.eventMan.Emit(ctx, event.Name(), paymentLink); emitErr != nil {
		logger.WithError(emitErr).Warn("could not emit payment link save event")
		return nil, emitErr
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

	if statusEmitErr := pb.eventMan.Emit(ctx, events.EventNameStatusSave, status); statusEmitErr != nil {
		logger.WithError(statusEmitErr).Warn("could not emit payment link status event")
		return nil, statusEmitErr
	}

	err = pb.service.Publish(ctx, "create.payment.link", paymentLink)
	if err != nil {
		logger.WithError(err).Warn("could not publish create-payment-link")
		// Emit the status event even if publish fails
		status.State = int32(commonv1.STATE_INACTIVE.Number())
		status.Status = int32(commonv1.STATUS_FAILED.Number())
		status.Extra["error"] = err.Error()
		if statusFailErr := pb.eventMan.Emit(ctx, events.EventNameStatusSave, status); statusFailErr != nil {
			logger.WithError(statusFailErr).Warn("could not emit payment link status event after publish failure")
		}
		return nil, err
	}
	return status.ToAPI(), nil
}

// validateAmountAndCost validates the amount and cost fields of the Payment.
func (pb *paymentBusiness) validateAmountAndCost(message *paymentv1.Payment, p *models.Payment, c *models.Cost) {
	if message.GetAmount().GetUnits() <= 0 || message.GetAmount().GetCurrencyCode() == "" {
		return
	}

	p.Amount = decimal.NullDecimal{
		Valid:   true,
		Decimal: utility.FromMoney(message.GetAmount()),
	}
	p.Currency = message.GetAmount().GetCurrencyCode()

	if message.GetCost().GetCurrencyCode() == "" {
		return
	}

	c.Amount = decimal.NullDecimal{
		Valid:   true,
		Decimal: utility.FromMoney(message.GetCost()),
	}
	c.Currency = message.GetCost().GetCurrencyCode()
}

const (
	// Transaction reference generation constants
	millionMod    = 1000000 // Modulo for time component
	asciiCharBase = 65      // ASCII A for character generation
	alphabetSize  = 26      // Number of letters in alphabet
	hundredKMod   = 100000  // Modulo for final component
)

// generateTransactionRef creates a unique 6-character reference for Jenga API.
func generateTransactionRef() string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	timeComponent := timestamp % millionMod
	asciiChar := asciiCharBase + ((timestamp / millionMod) % 26)
	return fmt.Sprintf("%c%05d", rune(asciiChar), timeComponent%hundredKMod)
}

// createDepositStep1 creates the initial receipt transaction:
// DR – Mobile Operator
// CR - Unidentified Deposits
func (pb *paymentBusiness) createDepositStep1(ctx context.Context, payment *models.Payment, senderTel, groupName, memberName string) error {
	if pb.ledgerCli == nil {
		return nil
	}
	logger := pb.service.Log(ctx).WithField("payment_id", payment.ID)

	// prepare account refs (pick consistent canonical refs)
	mobileOpRef := "mobile_operator"
	unidentifiedRef := "unidentified_deposits"

	// ensure accounts exist
	ledgerRef := "main_ledger" // adjust if you have ledger identifiers
	if err := pb.ensureLedgerAccount(ctx, mobileOpRef, ledgerRef, "mobile_operator"); err != nil {
		logger.WithError(err).Error("ensure mobile operator account")
		return err
	}
	if err := pb.ensureLedgerAccount(ctx, unidentifiedRef, ledgerRef, "suspense"); err != nil {
		logger.WithError(err).Error("ensure unidentified deposits account")
		return err
	}

	// amount as Money (reuse your utility)
	amount := utility.ToMoney(payment.Currency, payment.Amount.Decimal)

	// transaction reference (idempotency key)
	txRef := fmt.Sprintf("%s-deposit-step1", payment.ID)

	// OPTIONAL: idempotency check - depends on ledger API support
	// TODO: implement SearchTransactions or get by reference if ledger API has it.
	// if exists { log and return nil }

	// Build entries: DR MobileOperator, CR UnidentifiedDeposits
	entries := []*ledgerv1.TransactionEntry{
		{
			Account:      mobileOpRef,
			Transaction:  txRef,
			TransactedAt: time.Now().Format(time.RFC3339),
			Amount:       &amount,
			Credit:       false, // debit
		},
		{
			Account:      unidentifiedRef,
			Transaction:  txRef,
			TransactedAt: time.Now().Format(time.RFC3339),
			Amount:       &amount,
			Credit:       true, // credit
		},
	}

	extra := &data.JSONMap{
		"payment_id":   payment.ID,
		"payment_type": "DEPOSIT_STEP1",
		"narrative":    fmt.Sprintf("Funds deposited: %s", senderTel),
		"comments":     fmt.Sprintf("Funds deposited from %s", senderTel),
		"sender_tel":   senderTel,
		"member_name":  memberName,
		"group_name":   groupName,
		"original_ref": payment.ReferenceID,
	}

	// Create transaction
	transaction := &ledgerv1.Transaction{
		Reference:    txRef,
		Currency:     payment.Currency,
		TransactedAt: time.Now().Format(time.RFC3339),
		Data: extra.ToProtoStruct(),
		Entries: entries,
		Cleared: false,
		Type:    ledgerv1.TransactionType_NORMAL,
	}

	if _, err := pb.ledgerCli.CreateTransaction(ctx, connect.NewRequest(transaction)); err != nil {
		logger.WithError(err).Error("failed to create deposit step1 transaction")
		return err
	}
	logger.Info("created deposit step1 transaction (MobileOperator -> UnidentifiedDeposits)")
	return nil
}

// ensureLedgerAccount ensures that an account exists in the ledger service
func (pb *paymentBusiness) ensureLedgerAccount(ctx context.Context, accountRef, ledgerRef string, profileType string) error {
	if pb.ledgerCli == nil {
		return nil
	}

	logger := pb.service.Log(ctx).WithField("account_ref", accountRef)

	// Check if account already exists
	searchReq := &commonv1.SearchRequest{
		Query: fmt.Sprintf("reference:%s", accountRef),
	}

	accountStream, err := pb.ledgerCli.SearchAccounts(ctx, searchReq)
	if err != nil {
		logger.WithError(err).Error("failed to search for existing account")
		return err
	}

	// Try to receive one account to see if it exists
	_, err = accountStream.Recv()
	if err == nil {
		// Account exists, no need to create
		return nil
	}

	// Account doesn't exist, create it
	account := &ledgerv1.Account{
		Reference: accountRef,
		Ledger:    ledgerRef,
		Data: map[string]string{
			"profile_type": profileType,
			"created_by":   "payment_service",
		},
	}

	_, err = pb.ledgerCli.CreateAccount(ctx, account)
	if err != nil {
		logger.WithError(err).Error("failed to create ledger account")
		return err
	}

	logger.Info("successfully created ledger account")
	return nil
}


func (pb *paymentBusiness) Reconcile(ctx context.Context, msg *paymentv1.ReconcileRequest) (*paymentv1.ReconcileResponse, error) {
	// TODO implement me
	panic("implement me")
}