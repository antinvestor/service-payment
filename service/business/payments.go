package business

import (
	"context"
	"errors"

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
	Dispatch(ctx context.Context, payment *paymentV1.Payment) (*commonv1.StatusResponse, error)
	ReceivePayment(ctx context.Context, payment *paymentV1.Payment) (*commonv1.StatusResponse, error)
}

// validateDispatchFields checks for the required fields in the Payment model for dispatching
func (pb *paymentBusiness) validateDispatchFields(p *models.Payment) error {
	if p.SenderProfileID == "" || p.RecipientProfileID == "" {
		return ErrorPaymentDoesNotExist
	}
	return nil
}

// validateAmountAndCost validates the amount and cost fields of the Payment
func (pb *paymentBusiness) validateAmountAndCost(message *paymentV1.Payment, p *models.Payment, c *models.Cost) error {
	if message.GetAmount().Units == 0 || message.GetAmount().CurrencyCode == "" {
		return errors.New("amount is missing or invalid")
	}
	p.Amount = decimal.NullDecimal{
		Valid:   true,
		Decimal: decimal.NewFromFloat(float64(message.GetAmount().Units)),
	}
	p.Currency = message.GetAmount().CurrencyCode

	if message.GetCost().Units == 0 || message.GetCost().CurrencyCode == "" {
		return errors.New("cost is missing or invalid")
	}

	c.Amount = decimal.NullDecimal{
		Valid:   true,
		Decimal: decimal.NewFromFloat(float64(message.GetCost().Units)),
	}
	c.Currency = message.GetCost().CurrencyCode

	return nil
}

func (pb *paymentBusiness) emitPaymentEvent(ctx context.Context, p *models.Payment, c *models.Cost) error {
	event := events.PaymentSave{}
	if err := pb.service.Emit(ctx, event.Name(), p); err != nil {
		pb.service.L().WithError(err).Warn("could not emit payment event")
		return err
	}

	eventCost := events.CostSave{}
	if err := pb.service.Emit(ctx, eventCost.Name(), c); err != nil {
		pb.service.L().WithError(err).Warn("could not emit cost event")
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

func (pb *paymentBusiness) Dispatch(ctx context.Context, message *paymentV1.Payment) (*commonv1.StatusResponse, error) {
	logger := pb.service.L().WithField("request", message)

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
		Route:                message.GetRoute(),
		Outbound:             true,
	}

	//initialize cost

	c := &models.Cost{
		Amount: decimal.NullDecimal{
			Valid:   true,
			Decimal: decimal.NewFromFloat(float64(message.GetCost().Units)),
		},
		Currency: message.GetCost().CurrencyCode,
	}

	// Generate or validate Payment ID
	if message.GetId() == "" {
		p.GenID(ctx)
	}

	// Validate required fields
	if err := pb.validateDispatchFields(p); err != nil {
		logger.Error(err)
		return nil, err
	}

	// Validate and set amount and cost
	if err := pb.validateAmountAndCost(message, p, c); err != nil {
		logger.Error(err)
		return nil, err
	}

	// Set initial PaymentStatus,
	pStatus := models.PaymentStatus{
		PaymentID: message.GetId(),
		State:     int32(commonv1.STATE_CREATED.Number()),
		Status:    int32(commonv1.STATUS_QUEUED.Number()),
	}

	// Emit events for Payment and PaymentStatus
	if err := pb.emitPaymentEvent(ctx, p, c); err != nil {
		return nil, err
	}

	if err := pb.emitPaymentStatusEvent(ctx, pStatus); err != nil {
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}

func (pb *paymentBusiness) ReceivePayment(ctx context.Context, message *paymentV1.Payment) (*commonv1.StatusResponse, error) {
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
		Route:                message.GetRoute(),
		Outbound:             false,
	}

	c := &models.Cost{
		Amount: decimal.NullDecimal{
			Valid:   true,
			Decimal: decimal.NewFromFloat(float64(message.GetCost().Units)),
		},
		Currency: message.GetCost().CurrencyCode,
	}

	// Generate or validate Payment ID
	if message.GetId() == "" {
		p.GenID(ctx)
	}

	// Validate required fields
	if err := pb.validateReceiveFields(p); err != nil {
		logger.Error(err)
		return nil, err
	}

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
	if err := pb.emitPaymentEvent(ctx, p, c); err != nil {
		return nil, err
	}

	if err := pb.emitPaymentStatusEvent(ctx, pStatus); err != nil {
		return nil, err
	}

	return pStatus.ToStatusAPI(), nil
}

// validateReceiveFields checks for the required fields in the Payment model for receiving
func (pb *paymentBusiness) validateReceiveFields(p *models.Payment) error {
	if p.SenderProfileID == "" || p.RecipientProfileID == "" {
		return ErrorPaymentDoesNotExist
	}
	return nil
}
