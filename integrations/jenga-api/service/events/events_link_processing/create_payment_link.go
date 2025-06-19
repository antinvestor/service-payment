package events_link_processing

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/service/coreapi"
	models "github.com/antinvestor/jenga-api/service/models"
	"github.com/pitabwire/frame"
)

func GenerateExternalReference() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%013d", rand.Int63n(1e13))
}

// CreatePaymentLink handles the create.payment_link events
type CreatePaymentLink struct {
	Service       *frame.Service
	Client        coreapi.JengaApiClient
	PaymentClient *paymentV1.PaymentClient
}

// Name returns the name of the event handler
func (event *CreatePaymentLink) Name() string {
	return "create.payment.link"
}

// PayloadType returns the type of payload this event expects
func (event *CreatePaymentLink) PayloadType() any {
	return &models.PaymentLink{}
}

// Validate validates the payload
func (event *CreatePaymentLink) Validate(ctx context.Context, payload any) error {
	paymentLink, ok := payload.(*models.PaymentLink)
	if !ok {
		return fmt.Errorf("invalid payload type, expected PaymentLink")
	}

	// Basic validation
	if paymentLink.Name == "" {
		return fmt.Errorf("payment link name is required")
	}
	if paymentLink.Amount.IsZero() {
		return fmt.Errorf("payment link amount is required")
	}
	if paymentLink.ExpiryDate.IsZero() {
		return fmt.Errorf("expiry date is required")
	}
	if paymentLink.SaleDate.IsZero() {
		return fmt.Errorf("sale date is required")
	}
	if paymentLink.PaymentLinkType == "" {
		return fmt.Errorf("payment link type is required")
	}
	if paymentLink.SaleType == "" {
		return fmt.Errorf("sale type is required")
	}
	if paymentLink.AmountOption == "" {
		return fmt.Errorf("amount option is required")
	}
	if paymentLink.ExternalRef == "" {
		return fmt.Errorf("externalRef is required")
	}
	if paymentLink.Description == "" {
		return fmt.Errorf("description is required")
	}
	if paymentLink.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	return nil
}

// Handle implements the frame.SubscribeWorker interface
func (event *CreatePaymentLink) Handle(ctx context.Context, metadata map[string]string, message []byte) error {
	payload := event.PayloadType()
	paymentLink, ok := payload.(*models.PaymentLink)
	if !ok {
		return fmt.Errorf("invalid payload type, expected PaymentLink")
	}

	if err := json.Unmarshal(message, paymentLink); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	if err := event.Validate(ctx, paymentLink); err != nil {
		return fmt.Errorf("payload validation failed: %v", err)
	}

	return event.Execute(ctx, paymentLink)
}

// Execute handles the payment link creation logic
func (event *CreatePaymentLink) Execute(ctx context.Context, payload any) error {
	paymentLink, ok := payload.(*models.PaymentLink)
	if !ok {
		return fmt.Errorf("invalid payload type")
	}

	logger := event.Service.Log(ctx).WithField("paymentLinkId", paymentLink.ID)
	logger.Info("Processing create.payment_link event")

	// Prepare the request body for Jenga API
	var customers []models.PaymentLinkCustomer
	if err := json.Unmarshal(paymentLink.Customers, &customers); err != nil {
		logger.WithError(err).Error("failed to unmarshal customers")
		return fmt.Errorf("failed to unmarshal customers: %v", err)
	}

	var notifications []string
	if len(paymentLink.Notifications) > 0 {
		if err := json.Unmarshal(paymentLink.Notifications, &notifications); err != nil {
			logger.WithError(err).Error("failed to unmarshal notifications")
			return fmt.Errorf("failed to unmarshal notifications: %v", err)
		}
	}

	paymentLinkDetails := models.PaymentLinkDetails{
		ExpiryDate:      paymentLink.ExpiryDate.Format("2006-01-02"),
		SaleDate:        paymentLink.SaleDate.Format("2006-01-02"),
		PaymentLinkType: paymentLink.PaymentLinkType,
		SaleType:        paymentLink.SaleType,
		Name:            paymentLink.Name,
		Description:     paymentLink.Description,
		ExternalRef:     paymentLink.ExternalRef,
		PaymentLinkRef:  paymentLink.PaymentLinkRef,
		RedirectURL:     paymentLink.RedirectURL,
		AmountOption:    paymentLink.AmountOption,
		Amount:          paymentLink.Amount.InexactFloat64(),
		Currency:        paymentLink.Currency,
	}

	requestBody := models.PaymentLinkRequest{
		Customers:     customers,
		PaymentLink:   paymentLinkDetails,
		Notifications: notifications,
	}

	// Generate bearer token for authorization
	token, err := event.Client.GenerateBearerToken()
	if err != nil {
		logger.WithError(err).Error("failed to generate bearer token")
		statusUpdateRequest := &commonv1.StatusUpdateRequest{
			Id:     paymentLink.ID,
			State:  commonv1.STATE_ACTIVE,
			Status: commonv1.STATUS_FAILED,
			Extras: map[string]string{
				"update_type": "payment_link",
				"error":       fmt.Sprintf("failed to generate token: %v", err),
			},
		}
		_, updateErr := event.PaymentClient.StatusUpdate(ctx, statusUpdateRequest)
		if updateErr != nil {
			logger.WithError(updateErr).Error("failed to update payment link status")
		}
		return fmt.Errorf("failed to generate bearer token: %v", err)
	}

	// Make the API call to Jenga
	response, err := event.Client.CreatePaymentLink(requestBody, token.AccessToken)
	if err != nil || !response.Status {
		var errorMsg string
		if err != nil {
			errorMsg = err.Error()
		} else {
			errorMsg = fmt.Sprintf("API call failed with status: %v, message: %s", response.Status, response.Message)
		}
		logger.WithError(err).Error("failed to create payment link")
		statusUpdateRequest := &commonv1.StatusUpdateRequest{
			Id:     paymentLink.ID,
			State:  commonv1.STATE_ACTIVE,
			Status: commonv1.STATUS_FAILED,
			Extras: map[string]string{
				"update_type": "payment_link",
				"error":       errorMsg,
			},
		}
		_, updateErr := event.PaymentClient.StatusUpdate(ctx, statusUpdateRequest)
		if updateErr != nil {
			logger.WithError(updateErr).Error("failed to update payment link status")
		}
		//return fmt.Errorf("failed to create payment link: %v", errorMsg)
	}

	dataJson, err := json.Marshal(response.Data)
	if err != nil {
		logger.WithError(err).Error("failed to marshal payment link data")
	} else {
		logger.WithField("paymentLinkData", string(dataJson)).Info("Payment link data")
	}
	logger.WithField("response", response).Info("Payment link creation response received")

	// Update status to successful
	statusUpdateRequest := &commonv1.StatusUpdateRequest{
		Id:     paymentLink.ID,
		State:  commonv1.STATE_ACTIVE,
		Status: commonv1.STATUS_SUCCESSFUL,
		Extras: map[string]string{
			"update_type": "payment_link",
			"message":     "Payment link successfully generated",
		},
	}
	_, err = event.PaymentClient.StatusUpdate(ctx, statusUpdateRequest)
	if err != nil {
		logger.WithError(err).Error("failed to update payment link status")
	}

	return nil
}
