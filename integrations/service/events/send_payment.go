package events

import (
	"context"
	"context"
	"encoding/json"
	"fmt"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentv1 "github.com/antinvestor/apis/go/payment/v1"
	jengaApi "github.com/antinvestor/service-payments-v1/integrations/jenga-api/coreapi"
	"github.com/pitabwire/frame"

)

type SendPayment struct {
	Service *frame.Service
	JengaClient *jengaApi.Client
	PaymentClient *paymentv1.PaymentServiceClient

}

func (s *SendPayment)  Handle(ctx context.Context, _ map[string]string, payload []byte) error {
	log := ms.Service.L().WithField("type", "payment to send out")
	payment := &paymentv1.Payment{}
	err := json.Unmarshal(payload, payment)
	if err != nil {
		log.WithError(err).Error("could not unmarshal payment")
		return err
	}
	log.WithField("payment", payment).Info("got payment to send out")

	sender := client.ContactLink{}
	sender = sender.Populate(payment.GetSource())
	recipient := client.ContactLink{}
	recipient = recipient.Populate(payment.GetRecipient())

	paymentRequest := client.PaymentRequest{
		ID:          payment.GetId(),
		Sender:      sender,
		Recipient:   recipient,
		Amount:        payment.GetAmount(),
		TransactionId: payment.GetTransactionId(),
		ReferenceId:   payment.GetReferenceId(),
		Currency:      payment.GetCost().GetCurrencyCode(),
		PaymentType:   payment.GetPaymentType(),
		MetaData:      payment.GetMetaData(),
	}
	log.WithField("payment", paymentRequest).Info("got payment to send out")

	// check payment type use switch case
	switch paymentRequest.PaymentType {
	case "Bank Transfers":
		// send payment to jenga

	


