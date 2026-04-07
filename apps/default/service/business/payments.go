package business

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/ledger/connectrpc/go/v1/ledgerv1connect"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/profile/v1/profilev1connect"
	"buf.build/gen/go/antinvestor/tenancy/connectrpc/go/tenancy/v1/tenancyv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/default/service/events"
	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
	"github.com/pitabwire/frame/data"
	fevents "github.com/pitabwire/frame/events"
	"github.com/pitabwire/frame/queue"
	"github.com/pitabwire/frame/workerpool"
	"github.com/pitabwire/util"
	"github.com/pitabwire/util/decimalx"
	utilmoney "github.com/pitabwire/util/money"
)

func NewPaymentBusiness(
	_ context.Context,
	workMan workerpool.Manager,
	eventMan fevents.Manager,
	profileCli profilev1connect.ProfileServiceClient,
	tenancyCli tenancyv1connect.TenancyServiceClient,
	ledgerCli ledgerv1connect.LedgerServiceClient,
	paymentRepo repository.PaymentRepository,
	statusRepo repository.StatusRepository,
	costRepo repository.CostRepository,
	accountRepo repository.AccountRepository,
	promptRepo repository.PromptRepository,
	paymentLinkRepo repository.PaymentLinkRepository,
) (PaymentBusiness, error) {
	return &paymentBusiness{
		eventMan:        eventMan,
		profileCli:      profileCli,
		tenancyCli:      tenancyCli,
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
	qMan            queue.Manager
	eventMan        fevents.Manager
	profileCli      profilev1connect.ProfileServiceClient
	tenancyCli      tenancyv1connect.TenancyServiceClient
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

	costAmt := utilmoney.FromMoney(message.GetCost())
	c := &models.Cost{
		Amount:   costAmt.Ptr(),
		Currency: message.GetCost().GetCurrencyCode(),
	}
	c.GenID(ctx)

	if message.GetId() == "" {
		p.GenID(ctx)
	}

	pb.validateAmountAndCost(message, p, c)

	// Save cost separately and add its ID to payment
	if err := pb.eventMan.Emit(ctx, events.EventNameCostSave, c); err != nil {
		util.Log(ctx).WithError(err).Warn("could not emit cost event")
		return nil, err
	}
	p.CostIDs = []string{c.ID}

	if err := pb.eventMan.Emit(ctx, events.EventNamePaymentSave, p); err != nil {
		util.Log(ctx).WithError(err).Warn("could not emit payment event")
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
			var sourceExtras data.JSONMap
			sourceExtras = sourceExtras.FromProtoStruct(message.GetSource().GetExtras())
			memberName = sourceExtras.GetString("member_name")
		}
	}

	// try group name from payment-level extras first, then source extras
	groupName := ""
	if message.GetSource() != nil {
		var sourceExtras data.JSONMap
		sourceExtras = sourceExtras.FromProtoStruct(message.GetSource().GetExtras())
		groupName = sourceExtras.GetString("group_name")
	}

	// Create ledger transaction for outbound payment
	if pb.ledgerCli != nil && p.Amount != nil {
		if err := pb.createDepositStep1(ctx, p, senderTel, groupName, memberName); err != nil {
			util.Log(ctx).WithError(err).Warn("could not create ledger transaction for send")
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

	err := pb.statusRepo.Create(ctx, status)
	if err != nil {
		util.Log(ctx).WithError(err).Warn("could not save status")
		return nil, err
	}

	return status.ToAPI(), nil
}

func (pb *paymentBusiness) Receive(ctx context.Context, message *paymentv1.Payment) (*commonv1.StatusResponse, error) {
	logger := util.Log(ctx)
	logger.Debug("handling receive request")

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

	costAmt2 := utilmoney.FromMoney(message.GetCost())
	c := &models.Cost{
		Amount:   costAmt2.Ptr(),
		Currency: message.GetCost().GetCurrencyCode(),
	}
	c.GenID(ctx)

	if message.GetId() == "" {
		p.GenID(ctx)
	}
	pb.validateAmountAndCost(message, p, c)

	// Save cost separately and add its ID to payment
	if err := pb.eventMan.Emit(ctx, events.EventNameCostSave, c); err != nil {
		util.Log(ctx).WithError(err).Warn("could not emit cost event")
		return nil, err
	}
	p.CostIDs = []string{c.ID}

	if err := pb.eventMan.Emit(ctx, events.EventNamePaymentSave, p); err != nil {
		util.Log(ctx).WithError(err).Warn("could not emit payment event")
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
			var sourceExtras data.JSONMap
			sourceExtras = sourceExtras.FromProtoStruct(message.GetSource().GetExtras())
			memberName = sourceExtras.GetString("member_name")
		}
	}

	// try group name from payment-level extras first, then source extras
	groupName := ""
	if message.GetSource() != nil {
		var sourceExtras data.JSONMap
		sourceExtras = sourceExtras.FromProtoStruct(message.GetSource().GetExtras())
		groupName = sourceExtras.GetString("group_name")
	}

	// Create ledger transaction for inbound payment
	if pb.ledgerCli != nil && p.Amount != nil {
		if err := pb.createDepositStep1(ctx, p, senderTel, groupName, memberName); err != nil {
			util.Log(ctx).WithError(err).Warn("could not create ledger transaction for receive")
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
		util.Log(ctx).WithError(err).Warn("could not emit status event")
		return nil, err
	}

	return status.ToAPI(), nil
}

func (pb *paymentBusiness) Status(
	ctx context.Context,
	statusReq *commonv1.StatusRequest,
) (*commonv1.StatusResponse, error) {
	logger := util.Log(ctx).WithField("entity_id", statusReq.GetId())
	logger.Debug("handling status check request")

	var extras data.JSONMap
	extras = extras.FromProtoStruct(statusReq.GetExtras())
	entityType := extras.GetString("entity_type")

	status, err := pb.statusRepo.GetByEntity(ctx, statusReq.GetId(), entityType)
	if err != nil {
		return nil, err
	}
	return status.ToAPI(), nil
}

func (pb *paymentBusiness) StatusUpdate(
	ctx context.Context,
	req *commonv1.StatusUpdateRequest,
) (*commonv1.StatusResponse, error) {
	logger := util.Log(ctx).WithField("entity_id", req.GetId())
	logger.Debug("handling status update request")

	var extras data.JSONMap
	extras = extras.FromProtoStruct(req.GetExtras())

	entityType := extras.GetString("entity_type")
	if entityType == "" {
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

func (pb *paymentBusiness) convertPaymentsToAPI(
	ctx context.Context,
	paymentList []*models.Payment,
) ([]*paymentv1.Payment, error) {
	var responsesList []*paymentv1.Payment

	var paymentIDList []string

	for _, p := range paymentList {
		paymentIDList = append(paymentIDList, p.ID)
	}

	statusMap, err := pb.statusRepo.GetByEntityIDList(ctx, paymentIDList, "payment")
	if err != nil {
		return nil, err
	}

	for _, p := range paymentList {
		status := statusMap[p.ID]

		// Convert the payment model to the API response format
		result := p.ToAPI(status, nil)
		responsesList = append(responsesList, result)
	}

	return responsesList, nil
}

func (pb *paymentBusiness) Search(
	ctx context.Context,
	searchQuery *commonv1.SearchRequest,
) (workerpool.JobResultPipe[[]*paymentv1.Payment], error) {
	logger := util.Log(ctx)
	logger.Debug("handling payment search request")

	cursor := searchQuery.GetCursor()

	searchOpts := []data.SearchOption{
		data.WithSearchLimit(int(cursor.GetLimit())),
	}

	if searchQuery.GetIdQuery() != "" {
		searchOpts = append(
			searchOpts,
			data.WithSearchFiltersAndByValue(map[string]any{"id": searchQuery.GetIdQuery()}),
		)
	}

	if searchQuery.GetQuery() != "" {
		searchOpts = append(
			searchOpts,
			data.WithSearchFiltersOrByValue(
				map[string]any{"searchable @@ websearch_to_tsquery( 'english', ?) ": searchQuery.GetQuery()},
			),
		)

		for _, filter := range searchQuery.GetProperties() {
			searchOpts = append(
				searchOpts,
				data.WithSearchFiltersOrByValue(map[string]any{fmt.Sprintf(" %s = ?", filter): searchQuery.GetQuery()}),
			)
		}
	}

	query := data.NewSearchQuery(searchOpts...)
	results, err := pb.paymentRepo.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	processRes := workerpool.NewJob[[]*paymentv1.Payment](
		func(ctx context.Context, pipe workerpool.JobResultPipe[[]*paymentv1.Payment]) error {
			cancelCtx, cancel := context.WithCancel(ctx)
			defer cancel()

			for {
				res, ok := results.ReadResult(cancelCtx)
				if !ok {
					return nil
				}

				if res.IsError() {
					return res.Error()
				}

				finalRes, convErr := pb.convertPaymentsToAPI(cancelCtx, res.Item())
				if convErr != nil {
					return convErr
				}

				writeErr := pipe.WriteResult(cancelCtx, finalRes)
				if writeErr != nil {
					return writeErr
				}
			}
		},
	)

	return processRes, nil
}

func (pb *paymentBusiness) Release(
	ctx context.Context,
	paymentReq *paymentv1.ReleaseRequest,
) (*commonv1.StatusResponse, error) {
	logger := util.Log(ctx).WithField("payment_id", paymentReq.GetId())
	logger.Debug("handling release request")

	p, err := pb.paymentRepo.GetByID(ctx, paymentReq.GetId())
	if err != nil {
		logger.WithError(err).Warn("could not fetch payment by id")
		return nil, err
	}

	if !p.IsReleased() {
		releaseDate := time.Now()
		p.ReleasedAt = &releaseDate

		err = pb.eventMan.Emit(ctx, events.EventNamePaymentSave, p)
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

		return status.ToAPI(), nil
	}
	status, statusErr := pb.statusRepo.GetByEntity(ctx, p.ID, "payment")
	if statusErr != nil {
		logger.WithError(statusErr).Warn("could not get payment status")
		return nil, statusErr
	}
	return status.ToAPI(), nil
}

func (pb *paymentBusiness) InitiatePrompt(
	ctx context.Context,
	req *paymentv1.InitiatePromptRequest,
) (*commonv1.StatusResponse, error) {
	logger := util.Log(ctx)
	logger.Debug("handling initiate prompt request")

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

	var promptExtras data.JSONMap
	promptExtras = promptExtras.FromProtoStruct(req.GetExtra())

	p := &models.Prompt{
		SourceID:             req.GetSource().GetProfileId(),
		SourceProfileType:    req.GetSource().GetProfileType(),
		SourceContactID:      req.GetSource().GetContactId(),
		RecipientID:          req.GetRecipient().GetProfileId(),
		RecipientProfileType: req.GetRecipient().GetProfileType(),
		RecipientContactID:   req.GetRecipient().GetContactId(),
		Amount:               utilmoney.FromMoney(req.GetAmount()).Ptr(),
		DateCreated:          time.Now().Format("2006-01-02 15:04:05"),
		DeviceID:             req.GetDeviceId(),
		State:                int32(commonv1.STATE_CREATED.Number()),
		Status:               int32(commonv1.STATUS_QUEUED.Number()),
		AccountID:            accountPtr.ID,
		Account:              *accountPtr,
		Extra:                promptExtras,
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

	logger = logger.WithField("prompt_id", p.ID)
	logger.Debug("prompt ID set")

	p.Extra["transaction_ref"] = transactionRef
	p.Extra["currency"] = req.GetAmount().GetCurrencyCode()
	p.Extra["mobile_number"] = req.GetSource().GetDetail()
	// Add telco and pushType information if provided

	err = pb.eventMan.Emit(ctx, events.EventNamePromptSave, p)
	if err != nil {
		logger.WithError(err).Warn("could not emit prompt save")
		return nil, err
	}

	logger.Debug("prompt saved and event emitted for STK/USSD processing")

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

	err = pb.qMan.Publish(ctx, "initiate_prompt", p)
	if err != nil {
		logger.WithError(err).Warn("could not publish initiate-prompt")
		return nil, err
	}

	return status.ToAPI(), nil
}

//nolint:gocognit,funlen // Business logic complexity and length are acceptable for payment link creation
func (pb *paymentBusiness) CreatePaymentLink(
	ctx context.Context,
	req *paymentv1.CreatePaymentLinkRequest,
) (*commonv1.StatusResponse, error) {
	logger := util.Log(ctx)
	logger.Debug("handling create payment link request")

	// Validate required fields
	if req == nil || req.GetPaymentLink() == nil {
		logger.Error("missing payment link payload")
		return nil, errors.New("missing payment link payload")
	}

	plReq := req.GetPaymentLink()

	// Marshal customers to JSON
	var customersJSON data.JSONMap
	if len(req.GetCustomers()) > 0 {
		customers, err := pb.buildCustomersFromRequest(req.GetCustomers())
		if err != nil {
			logger.WithError(err).Error("failed to build customers")
			return nil, err
		}
		customersJSON = customers
	}

	// Marshal notifications to JSON
	var notificationsJSON data.JSONMap
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
		if unmarshalErr := json.Unmarshal(b, &notificationsJSON); unmarshalErr != nil {
			logger.WithError(unmarshalErr).Error("failed to convert notifications to JSONMap")
			return nil, unmarshalErr
		}
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
	amount := decimalx.Zero()
	if plReq.GetAmount() != nil {
		amount = utilmoney.FromMoney(plReq.GetAmount())
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
	if emitErr := pb.eventMan.Emit(ctx, events.EventNamePaymentLinkSave, paymentLink); emitErr != nil {
		logger.WithError(emitErr).Warn("could not emit payment link save event")
		return nil, emitErr
	}

	// Create PaymentLinkStatus
	status := &models.Status{
		EntityID:   paymentLink.ID,
		EntityType: "payment_link",
		State:      int32(commonv1.STATE_CREATED.Number()),
		Status:     int32(commonv1.STATUS_QUEUED.Number()),
		Extra:      make(data.JSONMap),
	}
	status.GenID(ctx)

	if statusEmitErr := pb.eventMan.Emit(ctx, events.EventNameStatusSave, status); statusEmitErr != nil {
		logger.WithError(statusEmitErr).Warn("could not emit payment link status event")
		return nil, statusEmitErr
	}

	err = pb.qMan.Publish(ctx, "create_payment_link", paymentLink)
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

	amt := utilmoney.FromMoney(message.GetAmount())
	p.Amount = &amt
	p.Currency = message.GetAmount().GetCurrencyCode()

	if message.GetCost().GetCurrencyCode() == "" {
		return
	}

	costVal := utilmoney.FromMoney(message.GetCost())
	c.Amount = &costVal
	c.Currency = message.GetCost().GetCurrencyCode()
}

const (
	// Transaction reference generation constants.
	millionMod    = 1000000 // Modulo for time component
	asciiCharBase = 65      // ASCII A for character generation
	alphabetSize  = 26      // Number of letters in alphabet
	hundredKMod   = 100000  // Modulo for final component
)

// generateTransactionRef creates a unique 6-character reference for Jenga API.
func generateTransactionRef() string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	timeComponent := timestamp % millionMod
	asciiChar := asciiCharBase + ((timestamp / millionMod) % alphabetSize)
	return fmt.Sprintf("%c%05d", rune(asciiChar), timeComponent%hundredKMod)
}

// createDepositStep1 creates the initial receipt transaction:
// DR – Mobile Operator
// CR - Unidentified Deposits.
func (pb *paymentBusiness) createDepositStep1(
	ctx context.Context,
	payment *models.Payment,
	senderTel, groupName, memberName string,
) error {
	if pb.ledgerCli == nil {
		return nil
	}
	logger := util.Log(ctx).WithField("payment_id", payment.ID)

	// prepare account refs (pick consistent canonical refs)
	mobileOpRef := "mobile_operator"
	unidentifiedRef := "unidentified_deposits"

	// ensure accounts exist
	ledgerRef := "main_ledger" // adjust if you have ledger identifiers
	if err := pb.ensureLedgerAccount(ctx, mobileOpRef, ledgerRef, "mobile_operator"); err != nil {
		return err
	}
	if err := pb.ensureLedgerAccount(ctx, unidentifiedRef, ledgerRef, "suspense"); err != nil {
		return err
	}

	// amount as Money (reuse money utility)
	amount := utilmoney.ToMoney(payment.Currency, decimalx.DerefOr(payment.Amount, decimalx.Zero()))

	// transaction reference (idempotency key)
	txRef := fmt.Sprintf("%s-deposit-step1", payment.ID)

	// OPTIONAL: idempotency check - depends on ledger API support
	// TODO: implement SearchTransactions or get by reference if ledger API has it.
	// if exists { log and return nil }

	// Build entries: DR MobileOperator, CR UnidentifiedDeposits
	entries := []*ledgerv1.TransactionEntry{
		{
			AccountId:     mobileOpRef,
			TransactionId: txRef,
			TransactedAt:  time.Now().Format(time.RFC3339),
			Amount:        amount,
			Credit:        false, // debit
		},
		{
			AccountId:     unidentifiedRef,
			TransactionId: txRef,
			TransactedAt:  time.Now().Format(time.RFC3339),
			Amount:        amount,
			Credit:        true, // credit
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
	transaction := &ledgerv1.CreateTransactionRequest{
		Id:           txRef,
		Currency:     payment.Currency,
		TransactedAt: time.Now().Format(time.RFC3339),
		Data:         extra.ToProtoStruct(),
		Entries:      entries,
		Cleared:      false,
		Type:         ledgerv1.TransactionType_NORMAL,
	}

	if _, err := pb.ledgerCli.CreateTransaction(ctx, connect.NewRequest(transaction)); err != nil {
		logger.WithError(err).Error("failed to create deposit step1 transaction")
		return err
	}
	logger.Debug("created deposit step1 transaction")
	return nil
}

// ensureLedgerAccount ensures that an account exists in the ledger service.
func (pb *paymentBusiness) ensureLedgerAccount(
	ctx context.Context,
	accountRef, ledgerRef string,
	profileType string,
) error {
	if pb.ledgerCli == nil {
		return nil
	}

	logger := util.Log(ctx).WithField("account_ref", accountRef)

	// Check if account already exists
	searchReq := &commonv1.SearchRequest{
		Query: fmt.Sprintf("reference:%s", accountRef),
	}

	accountStream, err := pb.ledgerCli.SearchAccounts(ctx, connect.NewRequest(searchReq))
	if err != nil {
		logger.WithError(err).Error("failed to search for existing account")
		return err
	}

	var accounts []*ledgerv1.Account
	for accountStream.Receive() {
		if accountStream.Err() != nil {
			return accountStream.Err()
		}

		accounts = append(accounts, accountStream.Msg().GetData()...)
	}

	if len(accounts) > 0 {
		logger.Debug("account already exists")
		return nil
	}

	accountData := data.JSONMap{
		"profile_type": profileType,
		"created_by":   "payment_service",
	}

	// Account doesn't exist, create it
	account := &ledgerv1.CreateAccountRequest{
		Id:       accountRef,
		Currency: "",
		LedgerId: ledgerRef,
		Data:     accountData.ToProtoStruct(),
	}

	_, err = pb.ledgerCli.CreateAccount(ctx, connect.NewRequest(account))
	if err != nil {
		logger.WithError(err).Error("failed to create ledger account")
		return err
	}

	logger.Debug("successfully created ledger account")
	return nil
}

// buildCustomersFromRequest builds customer models from the request.
func (pb *paymentBusiness) buildCustomersFromRequest(
	customers []*paymentv1.Customer,
) (data.JSONMap, error) {
	result := make([]models.Customer, 0, len(customers))
	for _, c := range customers {
		firstName, lastName := pb.splitProfileName(c.GetSource().GetProfileName())

		var customerExtras data.JSONMap
		customerExtras = customerExtras.FromProtoStruct(c.GetSource().GetExtras())

		result = append(result, models.Customer{
			FirstName:           firstName,
			LastName:            lastName,
			Email:               customerExtras.GetString("email"),
			PhoneNumber:         c.GetSource().GetContactId(),
			FirstAddress:        c.GetFirstAddress(),
			CountryCode:         c.GetCountryCode(),
			PostalOrZipCode:     c.GetPostalOrZipCode(),
			CustomerExternalRef: c.GetCustomerExternalRef(),
		})
	}
	// Convert to JSONMap via marshal/unmarshal
	bytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	var jsonMap data.JSONMap
	if unmarshalErr := json.Unmarshal(bytes, &jsonMap); unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return jsonMap, nil
}

// splitProfileName splits a profile name into first and last name.
func (pb *paymentBusiness) splitProfileName(profileName string) (string, string) {
	firstName := profileName
	lastName := ""
	if len(profileName) == 0 {
		return firstName, lastName
	}
	parts := strings.Fields(profileName)
	if len(parts) > 1 {
		firstName = parts[0]
		lastName = strings.Join(parts[1:], " ")
	} else {
		firstName = parts[0]
		lastName = ""
	}
	return firstName, lastName
}

func (pb *paymentBusiness) Reconcile(
	_ context.Context,
	_ *paymentv1.ReconcileRequest,
) (*paymentv1.ReconcileResponse, error) {
	// TODO implement me
	panic("implement me")
}
