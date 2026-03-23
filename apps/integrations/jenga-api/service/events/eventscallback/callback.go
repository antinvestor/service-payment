package eventscallback

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/payment/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/payment/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/pitabwire/frame/data"
	"github.com/pitabwire/util"
	"github.com/pitabwire/util/decimalx"
	utilmoney "github.com/pitabwire/util/money"
)

type JengaCallbackReceivePayment struct {
	paymentClient paymentv1connect.PaymentServiceClient
}

// NewJengaCallbackReceivePayment creates a new callback handler with dependencies.
func NewJengaCallbackReceivePayment(
	paymentClient paymentv1connect.PaymentServiceClient,
) *JengaCallbackReceivePayment {
	return &JengaCallbackReceivePayment{
		paymentClient: paymentClient,
	}
}

func (event *JengaCallbackReceivePayment) Name() string {
	return "jenga.callback.receive.payment"
}

func (event *JengaCallbackReceivePayment) PayloadType() any {
	return &models.CallbackRequest{}
}

func (event *JengaCallbackReceivePayment) Validate(_ context.Context, payload any) error {
	req, ok := payload.(*models.CallbackRequest)
	if !ok {
		return errors.New("invalid payload type")
	}

	if req.Transaction.Reference == "" {
		return errors.New("transaction reference is required")
	}

	return nil
}

func (event *JengaCallbackReceivePayment) Execute(ctx context.Context, payload any) error {
	logger := util.Log(ctx)

	req, ok := payload.(*models.CallbackRequest)
	if !ok {
		return errors.New("invalid payload type")
	}

	logger.WithField("callback", req).Info("Received Jenga callback for payment processing")

	callbackJSON, err := json.Marshal(req)
	if err != nil {
		logger.WithError(err).Error("failed to marshal callback")
		return nil
	}

	cbJSON := data.JSONMap{
		"additional_info": string(callbackJSON),
	}

	// Create base payment structure
	amtDec, _ := decimalx.NewFromString(fmt.Sprintf("%g", req.Transaction.Amount))
	amount := utilmoney.ToMoney(req.Transaction.Currency, amtDec)
	costDec, _ := decimalx.NewFromString(fmt.Sprintf("%g", req.Transaction.ServiceCharge))
	cost := utilmoney.ToMoney(req.Transaction.Currency, costDec)
	payment := &paymentv1.Payment{
		Source: &commonv1.ContactLink{
			Detail: req.Customer.MobileNumber,
		},
		Recipient: &commonv1.ContactLink{
			Detail: req.Bank.Account,
		},
		TransactionId: req.Transaction.Reference,
		Amount:        amount,
		Cost:          cost,
		Extra:         cbJSON.ToProtoStruct(),
	}

	receiveRequest := &paymentv1.ReceiveRequest{
		Data: payment,
	}

	// Invoke the Connect RPC receive method
	_, err = event.paymentClient.Receive(ctx, connect.NewRequest(receiveRequest))
	if err != nil {
		logger.WithError(err).Error("failed to receive payment")
		return err
	}
	return nil
}
