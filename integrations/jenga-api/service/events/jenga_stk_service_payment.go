package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/service/models"
	"google.golang.org/genproto/googleapis/type/money"
	"github.com/pitabwire/frame"
)

type JengaSTKServicePayment struct {
	Service       *frame.Service
	PaymentClient *paymentV1.PaymentClient
}

func (event *JengaSTKServicePayment) Name() string {
	return "jenga.stk.service.payment"
}

func (event *JengaSTKServicePayment) PayloadType() any {
	return &models.STKUSSDRequest{}
}

func (event *JengaSTKServicePayment) Validate(ctx context.Context, payload any) error {
	request := payload.(*models.STKUSSDRequest)

	if request.Merchant.AccountNumber == "" {
		return errors.New("merchant account number is required")
	}
	if request.Payment.Amount == "" {
		return errors.New("payment amount is required")
	}
	if request.Payment.MobileNumber == "" {
		return errors.New("mobile number is required")
	}

	return nil
}

func (event *JengaSTKServicePayment) Execute(ctx context.Context, payload any) error {
	if event.PaymentClient == nil {
		return errors.New("payment client not initialized")
	}

	stkRequest := payload.(*models.STKUSSDRequest)

	// Convert amount string to int64 (cents)
	amountFloat, err := strconv.ParseFloat(stkRequest.Payment.Amount, 64)
	if err != nil {
		return fmt.Errorf("invalid amount format: %v", err)
	}
	amountCents := int64(amountFloat * 100)

	// Extract relevant information from STK request
	payment := &paymentV1.Payment{
		ReferenceId: stkRequest.Payment.Ref,
		Amount: &money.Money{
			Units:        amountCents,
			CurrencyCode: stkRequest.Payment.Currency,
		},
		Source: &commonv1.ContactLink{
			ContactId: stkRequest.Payment.MobileNumber,
			Extras: map[string]string{
				"mobile_number": stkRequest.Payment.MobileNumber,
				"telco":         stkRequest.Payment.Telco,
				"push_type":     stkRequest.Payment.PushType,
			},
		},
		Recipient: &commonv1.ContactLink{
			ContactId: stkRequest.Merchant.AccountNumber,
			Extras: map[string]string{
				"account":       stkRequest.Merchant.AccountNumber,
				"country_code": stkRequest.Merchant.CountryCode,
				"name":         stkRequest.Merchant.Name,
			},
		},
	}

	// Add any additional information from STK response to extras
	extras := make(map[string]string)
	// Marshal the full request to JSON and store it in extras
	requestJSON, err := json.Marshal(stkRequest)
	if err == nil {
		extras["raw_stk_request"] = string(requestJSON)
	}

	payment.Extra = extras
	receiveRequest := &paymentV1.ReceiveRequest{
		Data: payment,
	}

	// Invoke the GRPC receive method
	_, err = event.PaymentClient.Client.Receive(ctx, receiveRequest)
	return err
}