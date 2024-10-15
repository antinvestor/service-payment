package business

import (
	"context"
	"github.com/antinvestor/service-payments-v1/service/repository"
	"time"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	partitionV1 "github.com/antinvestor/apis/go/partition/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	profileV1 "github.com/antinvestor/apis/go/profile/v1"
	"github.com/antinvestor/service-payments-v1/service/events"
	"github.com/antinvestor/service-payments-v1/service/models"
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
	logger := pb.service.L().WithField("request", message)

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

	if err := pb.validateAmountAndCost(message, p, c); err != nil {
		logger.Error(err)
		return nil, err
	}

	pStatus := models.PaymentStatus{
		PaymentID: message.GetId(),
		State:     int32(commonv1.STATE_CREATED.Number()),
		Status:    int32(commonv1.STATUS_QUEUED.Number()),
	}

	// Emit events for Payment and PaymentStatus
	if err := pb.emitPaymentEvent(ctx, p); err != nil {
		return nil, err
	}

	if err := pb.emitPaymentStatusEvent(ctx, pStatus); err != nil {
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}

func (pb *paymentBusiness) Receive(ctx context.Context, message *paymentV1.Payment) (*commonv1.StatusResponse, error) {
	logger := pb.service.L().WithField("request", message)
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

	// if p.SenderProfileID == "" || p.RecipientProfileID == "" {
	// 	return nil, nil
	// }

	// Validate and set amount and cost
	if err := pb.validateAmountAndCost(message, p, c); err != nil {
		logger.Error(err)
		return nil, err
	}

	// Set initial PaymentStatus
	pStatus := models.PaymentStatus{
		PaymentID: message.GetId(),
		State:     int32(commonv1.STATE_CREATED.Number()),
		Status:    int32(commonv1.STATUS_QUEUED.Number()),
	}

	// Emit events for Payment and PaymentStatus
	if err := pb.emitPaymentEvent(ctx, p); err != nil {
		return nil, err
	}

	if err := pb.emitPaymentStatusEvent(ctx, pStatus); err != nil {
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}

func (pb *paymentBusiness) Status(ctx context.Context, status *commonv1.StatusRequest) (*commonv1.StatusResponse, error) {

	logger := pb.service.L().WithField("request", status)
	logger.Info("handling status check request")

	paymentRepo := repository.NewPaymentRepository(ctx, pb.service)
	p, err := paymentRepo.GetByID(ctx, status.GetId())
	if err != nil {
		logger.WithError(err).Error("could not get payment status")
		return nil, err
	}

	paymentStatusRepo := repository.NewPaymentStatusRepository(ctx, pb.service)
	pStatus, err := paymentStatusRepo.GetByID(ctx, p.ID)
	if err != nil {
		logger.WithError(err).Error("could not get payment status")
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}

func (pb *paymentBusiness) StatusUpdate(ctx context.Context, req *commonv1.StatusUpdateRequest) (*commonv1.StatusResponse, error) {

	logger := pb.service.L().WithField("request", req)
	logger.Info("handling status update request")

	paymentRepo := repository.NewPaymentRepository(ctx, pb.service)
	p, err := paymentRepo.GetByID(ctx, req.GetId())
	if err != nil {
		logger.WithError(err).Error("could not get payment status")
		return nil, err
	}
	pStatus := models.PaymentStatus{
		PaymentID: p.ID,
		State:     int32(req.GetState()),
		Status:    int32(req.GetStatus()),
		Extra:     frame.DBPropertiesFromMap(req.GetExtras()),
	}

	pStatus.GenID(ctx)

	//queue out payment status for further processing
	eventStatus := events.PaymentStatusSave{}
	err = pb.service.Emit(ctx, eventStatus.Name(), pStatus)
	if err != nil {
		logger.WithError(err).Warn("could not save status")
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}

func (pb *paymentBusiness) Search(search *commonv1.SearchRequest,
	stream paymentV1.PaymentService_SearchServer) error {

	// Log the incoming search request
	logger := pb.service.L().WithField("request", search)
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

	logger := pb.service.L().WithField("request", paymentReq)
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


// validateAmountAndCost validates the amount and cost fields of the Payment
func (pb *paymentBusiness) validateAmountAndCost(message *paymentV1.Payment, p *models.Payment, c *models.Cost) error {
	if message.GetAmount().Units <= 0 || message.GetAmount().CurrencyCode == "" {
		return nil 
	}

	p.Amount = decimal.NullDecimal{
		Valid:   true,
		Decimal: decimal.NewFromFloat(float64(message.GetAmount().Units)),
	}
	p.Currency = message.GetAmount().CurrencyCode

	if message.GetCost().CurrencyCode == "" {
		return nil
	}

	c.Amount = decimal.NullDecimal{
		Valid:   true,
		Decimal: decimal.NewFromFloat(float64(message.GetCost().Units)),
	}
	c.Currency = message.GetCost().CurrencyCode

	return nil
}
func (pb *paymentBusiness) emitPaymentEvent(ctx context.Context, p *models.Payment) error {
	event := events.PaymentSave{}
	if err := pb.service.Emit(ctx, event.Name(), p); err != nil {
		pb.service.L().WithError(err).Warn("could not emit payment event")
		return err
	}


	return nil
}

func (pb *paymentBusiness) emitPaymentStatusEvent(ctx context.Context, pStatus models.PaymentStatus) error {
	eventStatus := events.PaymentStatusSave{}
	if err := pb.service.Emit(ctx, eventStatus.Name(), pStatus); err != nil {
		pb.service.L().WithError(err).Warn("could not emit payment status event")
		return err
	}
	return nil
}
