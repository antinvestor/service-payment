package events

import (
	"context"
	"encoding/json"
	"time"

	paymentv1 "github.com/antinvestor/apis/go/payment/v1"
	jengaApi "github.com/antinvestor/service-payments-v1/integrations/jenga-api/coreapi"
	"github.com/antinvestor/service-payments-v1/integrations/jenga-api/service/client"
	"github.com/pitabwire/frame"
	"github.com/shopspring/decimal"
	money "google.golang.org/genproto/googleapis/type/money"
)

type SendPayment struct {
	Service *frame.Service
	JengaClient *jengaApi.Client
	PaymentClient *paymentv1.PaymentClient

}


// MoneyToDecimal converts *money.Money to decimal.NullDecimal
func MoneyToDecimal(m *money.Money) decimal.NullDecimal {
	if m == nil {
		return decimal.NullDecimal{Valid: false}
	}
	amount := decimal.NewFromInt(m.Units).Add(decimal.NewFromInt(int64(m.Nanos)))
	return decimal.NullDecimal{
		Decimal: amount,
		Valid:   true,
	}
}


func (s *SendPayment)  Handle(ctx context.Context, _ map[string]string, payload []byte) (jengaApi.STKUSSDResponse, error) {
	log := s.Service.L(ctx).WithField("type", "payment to send out")
	payment := &paymentv1.Payment{}
	err := json.Unmarshal(payload, payment)
	if err != nil {
		log.WithError(err).Error("could not unmarshal payment")
		return jengaApi.STKUSSDResponse{}, err
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
		Amount:        MoneyToDecimal(payment.GetAmount()),
		TransactionId: payment.GetTransactionId(),
		ReferenceId:   payment.GetReferenceId(),
		Currency:      payment.Amount.GetCurrencyCode(),
		MetaData:      payment.GetExtra(),
	}
	log.WithField("payment", paymentRequest).Info("got payment to send out")
	
    //get the time now and format it to YYYY-MM-DD
	date := time.Now()
	date = date.UTC()

	//initiate STK/USSD push request
	request := jengaApi.STKUSSDRequest{
		Merchant: jengaApi.Merchant{
			CountryCode:   "KE",
			AccountNumber: s.JengaClient.MerchantCode,
			Name:          "Stawi",
		},
		Payment: jengaApi.Payment{
			Ref:         payment.GetId(),
			Amount:      MoneyToDecimal(payment.GetAmount()).Decimal.String(),
			Currency:    payment.Amount.GetCurrencyCode(),
			Telco:       "Safaricom",
			MobileNumber: recipient.ContactDetail,
			Date:        date.String(),
			CallBackUrl: "https://stawi.io",
			PushType:    "STK_PUSH",
		},
	}
	log.WithField("request", request).Info("initiating STK/USSD push request")
	//generate bearer token
	token, err := s.JengaClient.GenerateBearerToken()
	if err != nil {
		log.WithError(err).Error("could not generate bearer token")
		return jengaApi.STKUSSDResponse{} , err
	}

	response, err := s.JengaClient.InitiateSTKUSSD(request, token.AccessToken)

	if err != nil {
		log.WithError(err).Error("could not initiate STK/USSD push request")
		return jengaApi.STKUSSDResponse{} , err	
		}

	log.WithField("response", response).Info("initiated STK/USSD push request")

	// update status to SUCCESSFUL
	log.Debug("successfully sent out message")

	return *response, nil
		
	}

    


	


