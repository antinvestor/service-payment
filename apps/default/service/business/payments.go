// Copyright 2023-2026 Ant Investor Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package business

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/ledger/connectrpc/go/v2/v1/ledgerv1connect"
	ledgerv1 "buf.build/gen/go/antinvestor/ledger/protocolbuffers/go/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/v2/profile/v1/profilev1connect"
	"buf.build/gen/go/antinvestor/tenancy/connectrpc/go/v2/tenancy/v1/tenancyv1connect"
	"connectrpc.com/connect/v2"
	"github.com/antinvestor/service-payments/apps/default/service/contacts"
	"github.com/antinvestor/service-payments/apps/default/service/events"
	"github.com/antinvestor/service-payments/apps/default/service/metrics"
	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
	"github.com/pitabwire/frame/v3/data"
	fevents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/frame/v3/queue"
	"github.com/pitabwire/frame/v3/security"
	"github.com/pitabwire/frame/v3/workerpool"
	"github.com/pitabwire/util"
	"github.com/pitabwire/util/decimalx"
	utilmoney "github.com/pitabwire/util/moneyx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

func NewPaymentBusiness(
	_ context.Context,
	workMan workerpool.Manager,
	qMan queue.Manager,
	promptTopicName string,
	paymentLinkTopicName string,
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
		qMan:                 qMan,
		promptTopicName:      promptTopicName,
		paymentLinkTopicName: paymentLinkTopicName,
		eventMan:             eventMan,
		profileCli:           profileCli,
		tenancyCli:           tenancyCli,
		ledgerCli:            ledgerCli,
		paymentRepo:          paymentRepo,
		statusRepo:           statusRepo,
		costRepo:             costRepo,
		accountRepo:          accountRepo,
		promptRepo:           promptRepo,
		paymentLinkRepo:      paymentLinkRepo,
		workMan:              workMan,
		paymentMetrics:       metrics.NewPaymentMetrics(),
		obs:                  metrics.NewMetrics(),
	}, nil
}

type paymentBusiness struct {
	qMan                 queue.Manager
	promptTopicName      string
	paymentLinkTopicName string
	eventMan             fevents.Manager
	profileCli           profilev1connect.ProfileServiceClient
	tenancyCli           tenancyv1connect.TenancyServiceClient
	ledgerCli            ledgerv1connect.LedgerServiceClient
	paymentRepo          repository.PaymentRepository
	statusRepo           repository.StatusRepository
	costRepo             repository.CostRepository
	accountRepo          repository.AccountRepository
	promptRepo           repository.PromptRepository
	paymentLinkRepo      repository.PaymentLinkRepository
	workMan              workerpool.Manager
	paymentMetrics       *metrics.PaymentMetrics
	obs                  *metrics.Metrics
}

//nolint:nonamedreturns // named err captured by deferred span-end closure
func (pb *paymentBusiness) Send(
	ctx context.Context,
	message *paymentv1.Payment,
) (resp *commonv1.StatusResponse, err error) {
	ctx, span := pb.obs.StartSpan(ctx, "Send",
		trace.WithAttributes(
			attribute.String("payment.route", message.GetRoute()),
			attribute.Bool("payment.outbound", true),
		),
	)
	defer func() { pb.obs.EndSpan(ctx, span, err) }()

	// A reference names one payout: sending it again answers the first
	// payment's status instead of paying twice.
	if prior, priorErr := pb.priorSend(ctx, message); prior != nil || priorErr != nil {
		return prior, priorErr
	}

	// The payee's contact decides how the money leaves: the route is chosen
	// by the payment's partition and the kind of contact paid. A route the
	// caller names is ignored. Only the contact id is stored; the number is read again as the
	// payment leaves for its provider.
	payee, err := contacts.ResolvePayee(ctx, pb.profileCli, message.GetRecipient())
	switch {
	case errors.Is(err, contacts.ErrUnknown):
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
	case errors.Is(err, contacts.ErrNoMobileNumber):
		return nil, connect.NewError(connect.CodeFailedPrecondition, err.Error()).WithCause(err)
	case err != nil:
		return nil, err
	}
	recipientContactID := message.GetRecipient().GetContactId()
	if recipientContactID == "" {
		recipientContactID = payee.ID
	}

	p := &models.Payment{
		SenderProfileType:    message.GetSource().GetProfileType(),
		SenderProfileID:      message.GetSource().GetProfileId(),
		SenderContactID:      message.GetSource().GetContactId(),
		RecipientProfileType: message.GetRecipient().GetProfileType(),
		RecipientProfileID:   message.GetRecipient().GetProfileId(),
		RecipientContactID:   recipientContactID,
		ReferenceID:          message.GetReferenceId(),
		BatchID:              message.GetBatchId(),
		PaymentType:          payee.Kind,
		OutBound:             true,
	}

	costAmt := utilmoney.FromMoney(message.GetCost())
	c := &models.Cost{
		Amount:   costAmt.Ptr(),
		Currency: message.GetCost().GetCurrencyCode(),
	}
	// Under a reference the ids are derived from it, so two sends racing
	// past priorSend save one payment, one cost and one status.
	referenceID := referencedID(ctx, message.GetReferenceId())
	if referenceID != "" {
		p.ID = referenceID
		c.ID = referenceID + "-cost"
	}
	c.GenID(ctx)

	p.GenID(ctx)

	pb.validateAmountAndCost(message, p, c)

	status := &models.Status{
		EntityID:   p.GetID(),
		EntityType: "payment",
		State:      int32(commonv1.STATE_CREATED.Number()),
		Status:     int32(commonv1.STATUS_QUEUED.Number()),
		Extra:      make(data.JSONMap),
	}
	if referenceID != "" {
		// The queued status is the reference's claim, saved before
		// anything is sent on.
		status.ID = referenceID + "-queued"
		if claimed, claimErr := pb.claimReference(ctx, message, status); claimed != nil || claimErr != nil {
			return claimed, claimErr
		}
	}

	// Save cost separately and add its ID to payment
	if err = pb.eventMan.Emit(ctx, events.EventNameCostSave, c); err != nil {
		util.Log(ctx).WithError(err).Warn("could not emit cost event")
		return nil, err
	}
	p.CostIDs = []string{c.ID}

	if err = pb.eventMan.Emit(ctx, events.EventNamePaymentSave, p); err != nil {
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
		if ledgerErr := pb.createDepositStep1(ctx, p, senderTel, groupName, memberName); ledgerErr != nil {
			util.Log(ctx).WithError(ledgerErr).Warn("could not create ledger transaction for send")
			// Don't fail the payment if ledger fails, just log the error
		}
	}

	if referenceID == "" {
		status.GenID(ctx)
		if err = pb.statusRepo.Create(ctx, status); err != nil {
			util.Log(ctx).WithError(err).Warn("could not save status")
			return nil, err
		}
	}

	pb.paymentMetrics.RecordInitiated(ctx, p)

	return status.ToAPI(), nil
}

// claimReference saves the queued status a referenced send starts with. Its
// id comes from the reference, so of two sends racing past priorSend only
// one saves it. The other answers the saved payment, checked as priorSend
// checks it; while that payment is still on its way to the database, it goes
// on and saves the same payment under the same ids, which the save event
// stores once. It returns nil, nil when the caller should go on sending.
func (pb *paymentBusiness) claimReference(
	ctx context.Context,
	message *paymentv1.Payment,
	status *models.Status,
) (*commonv1.StatusResponse, error) {
	status.GenID(ctx)
	err := pb.statusRepo.Create(ctx, status)
	if err == nil {
		return nil, nil
	}
	if !data.ErrorIsDuplicateKey(err) {
		util.Log(ctx).WithError(err).Warn("could not save status")
		return nil, err
	}
	return pb.priorSend(ctx, message)
}

// maxReferenceLength is the longest reference a payment row stores.
const maxReferenceLength = 150

// ErrReferenceReused is returned when a reference already names a payout to
// someone else or of another amount.
var ErrReferenceReused = errors.New("reference_id already names a different payment")

// referencedID derives a payment id from the caller's tenancy and reference,
// "" when there is no reference.
func referencedID(ctx context.Context, reference string) string {
	if reference == "" {
		return ""
	}
	claims := security.ClaimsFromContext(ctx)
	sum := sha256.Sum256([]byte(claims.GetTenantID() + "/" + claims.GetPartitionID() + "/" + reference))
	return "ref" + hex.EncodeToString(sum[:])[:32]
}

// priorSend answers the status of the outbound payment already sent under
// the message's reference, nil when there is none. The earlier payment must
// be to the same recipient for the same amount. A reference too long to
// store is refused here rather than lost when the payment is saved.
func (pb *paymentBusiness) priorSend(ctx context.Context, message *paymentv1.Payment) (*commonv1.StatusResponse, error) {
	reference := message.GetReferenceId()
	if reference == "" {
		return nil, nil
	}
	if len(reference) > maxReferenceLength {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			fmt.Sprintf("reference_id is longer than %d characters", maxReferenceLength))
	}
	claims := security.ClaimsFromContext(ctx)
	prior, err := pb.paymentRepo.GetOutboundByReference(ctx,
		claims.GetTenantID(), claims.GetPartitionID(), reference)
	if data.ErrorIsNoRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("look up payment by reference: %w", err)
	}
	amount := utilmoney.FromMoney(message.GetAmount())
	if prior.RecipientProfileID != message.GetRecipient().GetProfileId() ||
		prior.Currency != message.GetAmount().GetCurrencyCode() ||
		prior.Amount == nil || !prior.Amount.Equal(amount) {
		return nil, connect.NewError(connect.CodeAlreadyExists, ErrReferenceReused.Error())
	}
	status, err := pb.statusRepo.GetByEntity(ctx, prior.GetID(), "payment")
	if err != nil {
		return nil, fmt.Errorf("read status of payment %s: %w", prior.GetID(), err)
	}
	return status.ToAPI(), nil
}

//nolint:nonamedreturns // named err captured by deferred span-end closure
func (pb *paymentBusiness) Receive(
	ctx context.Context,
	message *paymentv1.Payment,
) (resp *commonv1.StatusResponse, err error) {
	ctx, span := pb.obs.StartSpan(ctx, "Receive",
		trace.WithAttributes(
			attribute.String("payment.route", message.GetRoute()),
			attribute.Bool("payment.outbound", false),
		),
	)
	defer func() { pb.obs.EndSpan(ctx, span, err) }()

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
		TransactionID:        message.GetTransactionId(),
		OutBound:             false,
	}
	p.Extra = p.Extra.FromProtoStruct(message.GetExtra())

	costAmt2 := utilmoney.FromMoney(message.GetCost())
	c := &models.Cost{
		Amount:   costAmt2.Ptr(),
		Currency: message.GetCost().GetCurrencyCode(),
	}
	c.GenID(ctx)

	// A caller-chosen id makes a retried receive the same payment.
	p.ID = message.GetId()
	if p.ID == "" {
		p.GenID(ctx)
	}
	pb.validateAmountAndCost(message, p, c)

	// Save cost separately and add its ID to payment
	if err = pb.eventMan.Emit(ctx, events.EventNameCostSave, c); err != nil {
		util.Log(ctx).WithError(err).Warn("could not emit cost event")
		return nil, err
	}
	p.CostIDs = []string{c.ID}

	if err = pb.eventMan.Emit(ctx, events.EventNamePaymentSave, p); err != nil {
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
		if ledgerErr := pb.createDepositStep1(ctx, p, senderTel, groupName, memberName); ledgerErr != nil {
			util.Log(ctx).
				WithError(ledgerErr).
				Warn("could not create ledger transaction for receive")

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

	if err = pb.eventMan.Emit(ctx, events.EventNameStatusSave, status); err != nil {
		util.Log(ctx).WithError(err).Warn("could not emit status event")
		return nil, err
	}

	pb.paymentMetrics.RecordInitiated(ctx, p)

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

//nolint:nonamedreturns // named err captured by deferred span-end closure
func (pb *paymentBusiness) StatusUpdate(
	ctx context.Context,
	req *commonv1.StatusUpdateRequest,
) (resp *commonv1.StatusResponse, err error) {
	ctx, span := pb.obs.StartSpan(ctx, "StatusUpdate")
	defer func() { pb.obs.EndSpan(ctx, span, err) }()

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

	if err = pb.eventMan.Emit(ctx, events.EventNameStatusSave, status); err != nil {
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
				map[string]any{
					"searchable @@ websearch_to_tsquery( 'english', ?) ": searchQuery.GetQuery(),
				},
			),
		)

		for _, filter := range searchQuery.GetProperties() {
			searchOpts = append(
				searchOpts,
				data.WithSearchFiltersOrByValue(
					map[string]any{fmt.Sprintf(" %s = ?", filter): searchQuery.GetQuery()},
				),
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

//nolint:nonamedreturns // named err captured by deferred span-end closure
func (pb *paymentBusiness) Release(
	ctx context.Context,
	paymentReq *paymentv1.ReleaseRequest,
) (resp *commonv1.StatusResponse, err error) {
	ctx, span := pb.obs.StartSpan(ctx, "Release",
		trace.WithAttributes(attribute.String("payment.id", paymentReq.GetId())),
	)
	defer func() { pb.obs.EndSpan(ctx, span, err) }()

	logger := util.Log(ctx).WithField("payment_id", paymentReq.GetId())
	logger.Debug("handling release request")

	var p *models.Payment
	p, err = pb.paymentRepo.GetByID(ctx, paymentReq.GetId())
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
	var status *models.Status
	status, err = pb.statusRepo.GetByEntity(ctx, p.ID, "payment")
	if err != nil {
		logger.WithError(err).Warn("could not get payment status")
		return nil, err
	}
	return status.ToAPI(), nil
}

//nolint:nonamedreturns // named err captured by deferred span-end closure
func (pb *paymentBusiness) InitiatePrompt(
	ctx context.Context,
	req *paymentv1.InitiatePromptRequest,
) (resp *commonv1.StatusResponse, err error) {
	ctx, span := pb.obs.StartSpan(ctx, "InitiatePrompt")
	start := time.Now()
	defer func() {
		pb.obs.EndSpan(ctx, span, err)
		if err != nil {
			pb.obs.RecordPromptFailed(ctx)
		} else {
			pb.obs.RecordPromptInitiated(ctx)
			pb.obs.RecordPromptLatency(ctx, time.Since(start))
		}
	}()

	logger := util.Log(ctx)
	logger.Debug("handling initiate prompt request")

	// A caller-chosen id names one prompt: asking again returns where that
	// prompt stands instead of pushing the payer a second time.
	if req.GetId() != "" {
		if existing, getErr := pb.promptRepo.GetByID(ctx, req.GetId()); getErr == nil && existing != nil {
			logger.WithField("prompt_id", req.GetId()).Debug("prompt already raised; not sent again")
			return existing.ToAPIStatus(), nil
		}
	}

	// Resolve the payer's number before anything is stored, so a prompt
	// that cannot be dialled is refused rather than left queued.
	var dial string
	if dial, err = payerNumber(ctx, pb.profileCli, req.GetSource()); err != nil {
		return nil, err
	}

	// Build Account from request
	account := models.Account{
		AccountNumber: req.GetRecipientAccount().GetAccountNumber(),
		CountryCode:   req.GetRecipientAccount().GetCountryCode(),
		Name:          req.GetRecipientAccount().GetName(),
	}

	// Use AccountRepository to get or create the account
	var accountPtr *models.Account
	accountPtr, err = pb.accountRepo.GetByAccountNumber(ctx, account.AccountNumber)
	if err != nil {
		// If not found, create the account
		account.GenID(ctx)
		err = pb.eventMan.Emit(ctx, events.EventNameAccountSave, &account)
		if err != nil {
			logger.WithError(err).Warn("could not create account")
			return nil, err
		}
		accountPtr = &account
		err = nil
	}

	transactionRef := generateTransactionRef()
	p := buildPromptModel(ctx, req, accountPtr, transactionRef)

	logger = logger.WithField("prompt_id", p.ID)
	logger.Debug("prompt ID set")

	err = pb.eventMan.Emit(ctx, events.EventNamePromptSave, p)
	if err != nil {
		logger.WithError(err).Warn("could not emit prompt save")
		return nil, err
	}

	logger.Debug("prompt saved and event emitted for STK/USSD processing")

	status, err := pb.publishPromptStatus(ctx, p.ID, transactionRef)
	if err != nil {
		return nil, err
	}

	// Publish InitiatePromptRequest protobuf (not the ORM model) so integrators
	// can proto.Unmarshal. Route-specific publishers take priority when registered.
	apiPrompt := p.ToAPI(nil)
	if dial != "" {
		// Providers dial the source contact; only the published prompt
		// carries the number, the stored prompt keeps the contact id.
		apiPrompt.Source.ContactId = dial
	}
	binaryPrompt, marshalErr := proto.Marshal(apiPrompt)
	if marshalErr != nil {
		logger.WithError(marshalErr).Warn("could not marshal initiate-prompt")
		return nil, marshalErr
	}

	topicName := pb.resolvePromptTopic(p.Route)
	logger = logger.WithField("prompt_topic", topicName).WithField("route", p.Route)
	err = pb.qMan.Publish(ctx, topicName, binaryPrompt)
	if err != nil {
		logger.WithError(err).Warn("could not publish initiate-prompt")
		return nil, err
	}

	return status.ToAPI(), nil
}

// resolvePromptTopic picks the queue publisher for a checkout/payment route.
// Prefers "prompt.<route>" when registered via INITIATE_PROMPT_ROUTE_URIS;
// otherwise uses the default INITIATE_PROMPT_TOPIC_NAME.
func (pb *paymentBusiness) resolvePromptTopic(route string) string {
	route = strings.TrimSpace(strings.ToLower(route))
	if route == "" {
		return pb.promptTopicName
	}
	ref := promptRoutePublisherRef(route)
	if pb.qMan != nil {
		if _, err := pb.qMan.GetPublisher(ref); err == nil {
			return ref
		}
	}
	return pb.promptTopicName
}

// promptRoutePublisherRef is the stable publisher name for a route key.
func promptRoutePublisherRef(route string) string {
	return "prompt." + strings.TrimSpace(strings.ToLower(route))
}

// buildPromptModel constructs a Prompt model from the initiate-prompt request,
// attaching the resolved account and seeding Extra fields.
func buildPromptModel(
	ctx context.Context,
	req *paymentv1.InitiatePromptRequest,
	accountPtr *models.Account,
	transactionRef string,
) *models.Prompt {
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
		Route:                req.GetRoute(),
		AccountID:            accountPtr.ID,
		Account:              *accountPtr,
		Extra:                promptExtras,
	}

	if req.GetId() != "" {
		p.ID = req.GetId()
	}
	if p.ID == "" {
		p.GenID(ctx)
		p.ID = p.GetID()
	}

	p.Extra["transaction_ref"] = transactionRef
	p.Extra["currency"] = req.GetAmount().GetCurrencyCode()
	p.Extra["mobile_number"] = req.GetSource().GetDetail()
	return p
}

// publishPromptStatus saves a queued Status record for the prompt and returns it.
func (pb *paymentBusiness) publishPromptStatus(
	ctx context.Context,
	promptID, transactionRef string,
) (*models.Status, error) {
	status := &models.Status{
		EntityID:   promptID,
		EntityType: "prompt",
		State:      int32(commonv1.STATE_CREATED.Number()),
		Status:     int32(commonv1.STATUS_QUEUED.Number()),
		Extra:      make(data.JSONMap),
	}
	status.GenID(ctx)
	status.Extra["transaction_ref"] = transactionRef

	if err := pb.eventMan.Emit(ctx, events.EventNameStatusSave, status); err != nil {
		util.Log(ctx).WithError(err).Warn("could not save status")
		return nil, err
	}
	return status, nil
}

//nolint:gocognit,funlen,nonamedreturns // Business logic complexity and length are acceptable for payment link creation; named err captured by deferred span-end closure
func (pb *paymentBusiness) CreatePaymentLink(
	ctx context.Context,
	req *paymentv1.CreatePaymentLinkRequest,
) (resp *commonv1.StatusResponse, err error) {
	ctx, span := pb.obs.StartSpan(ctx, "CreatePaymentLink")
	defer func() { pb.obs.EndSpan(ctx, span, err) }()

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
		customers, buildErr := pb.buildCustomersFromRequest(req.GetCustomers())
		if buildErr != nil {
			logger.WithError(buildErr).Error("failed to build customers")
			return nil, buildErr
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
		b, marshalErr := json.Marshal(notificationTypes)
		if marshalErr != nil {
			logger.WithError(marshalErr).Error("failed to marshal notifications")
			return nil, marshalErr
		}
		if unmarshalErr := json.Unmarshal(b, &notificationsJSON); unmarshalErr != nil {
			logger.WithError(unmarshalErr).Error("failed to convert notifications to JSONMap")
			return nil, unmarshalErr
		}
	}

	// Parse dates
	expiryDate, parseErr := time.Parse("2006-01-02", plReq.GetExpiryDate())
	if parseErr != nil {
		expiryDate = time.Now().Add(1 * 24 * time.Hour) // default: 1 days from now
	}
	saleDate, parseErr2 := time.Parse("2006-01-02", plReq.GetSaleDate())
	if parseErr2 != nil {
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

	err = pb.qMan.Publish(ctx, pb.paymentLinkTopicName, paymentLink)
	if err != nil {
		logger.WithError(err).Warn("could not publish create-payment-link")
		// Emit the status event even if publish fails
		status.State = int32(commonv1.STATE_INACTIVE.Number())
		status.Status = int32(commonv1.STATUS_FAILED.Number())
		status.Extra["error"] = err.Error()
		if statusFailErr := pb.eventMan.Emit(ctx, events.EventNameStatusSave, status); statusFailErr != nil {
			logger.WithError(statusFailErr).
				Warn("could not emit payment link status event after publish failure")
		}
		return nil, err
	}
	pb.obs.RecordPaymentLinkCreated(ctx)
	return status.ToAPI(), nil
}

// validateAmountAndCost validates the amount and cost fields of the Payment.
func (pb *paymentBusiness) validateAmountAndCost(
	message *paymentv1.Payment,
	p *models.Payment,
	c *models.Cost,
) {
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

	if _, err := pb.ledgerCli.CreateTransaction(ctx, transaction); err != nil {
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

	accountStream, err := pb.ledgerCli.SearchAccounts(ctx, searchReq)
	if err != nil {
		logger.WithError(err).Error("failed to search for existing account")
		return err
	}
	defer func() { _ = accountStream.Close() }()

	var accounts []*ledgerv1.Account
	for {
		msg, recvErr := accountStream.Receive()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			return recvErr
		}

		accounts = append(accounts, msg.GetData()...)
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

	_, err = pb.ledgerCli.CreateAccount(ctx, account)
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

//nolint:nonamedreturns // named err captured by deferred span-end closure
func (pb *paymentBusiness) Reconcile(
	ctx context.Context,
	_ *paymentv1.ReconcileRequest,
) (resp *paymentv1.ReconcileResponse, err error) {
	ctx, span := pb.obs.StartSpan(ctx, "Reconcile")
	defer func() { pb.obs.EndSpan(ctx, span, err) }()

	return nil, connect.NewError(
		connect.CodeUnimplemented,
		"reconcile not yet implemented",
	)
}
