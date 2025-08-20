//nolint:revive // package name matches directory structure
package events_link_processing //nolint:staticcheck // underscore package name required by project structure

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"

	commonv1 "github.com/antinvestor/apis/go/common/v1"
	paymentV1 "github.com/antinvestor/apis/go/payment/v1"
	"github.com/antinvestor/jenga-api/service/coreapi"
	models "github.com/antinvestor/jenga-api/service/models"
	"github.com/pitabwire/frame"
)

const (
	dateFormat            = "2006-01-02"
	updateTypePaymentLink = "payment_link"
	statusActive          = commonv1.STATE_ACTIVE
	statusFailed          = commonv1.STATUS_FAILED
	statusSuccessful      = commonv1.STATUS_SUCCESSFUL
	fallbackExternalRef   = "0000000000000"
	maxExternalRefValue   = 1e13
	externalRefLength     = 13
)

func GenerateExternalReference() string {
	var n uint64
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fallbackExternalRef
	}
	n = uint64(b[0])<<40 | uint64(b[1])<<32 | uint64(b[2])<<24 |
		uint64(b[3])<<16 | uint64(b[4])<<8 | uint64(b[5])
	n = n % maxExternalRefValue
	return fmt.Sprintf("%0*d", externalRefLength, n)
}

type CreatePaymentLink struct {
	Service       *frame.Service
	Client        coreapi.JengaApiClient
	PaymentClient paymentV1.PaymentClient
}

func NewCreatePaymentLink(service *frame.Service, client coreapi.JengaApiClient,
	paymentClient paymentV1.PaymentClient) *CreatePaymentLink {
	return &CreatePaymentLink{
		Service:       service,
		Client:        client,
		PaymentClient: paymentClient,
	}
}

func (h *CreatePaymentLink) Name() string {
	return "create.payment.link"
}

func (h *CreatePaymentLink) PayloadType() any {
	return &models.PaymentLink{}
}

func (h *CreatePaymentLink) Validate(ctx context.Context, payload any) error {
	paymentLink, ok := payload.(*models.PaymentLink)
	if !ok {
		return errors.New("invalid payload type, expected *models.PaymentLink")
	}

	switch {
	case paymentLink.Name == "":
		return errors.New("payment link name is required")
	case paymentLink.Amount.IsZero():
		return errors.New("payment link amount is required")
	case paymentLink.ExpiryDate.IsZero():
		return errors.New("expiry date is required")
	case paymentLink.SaleDate.IsZero():
		return errors.New("sale date is required")
	case paymentLink.PaymentLinkType == "":
		return errors.New("payment link type is required")
	case paymentLink.SaleType == "":
		return errors.New("sale type is required")
	case paymentLink.AmountOption == "":
		return errors.New("amount option is required")
	case paymentLink.ExternalRef == "":
		return errors.New("externalRef is required")
	case paymentLink.Description == "":
		return errors.New("description is required")
	case paymentLink.Currency == "":
		return errors.New("currency is required")
	}
	return nil
}

func (h *CreatePaymentLink) Handle(ctx context.Context, metadata map[string]string, message []byte) error {
	payload := h.PayloadType()
	if err := json.Unmarshal(message, payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	if err := h.Validate(ctx, payload); err != nil {
		return fmt.Errorf("payload validation failed: %w", err)
	}
	return h.Execute(ctx, payload)
}

func (h *CreatePaymentLink) Execute(ctx context.Context, payload any) error {
	paymentLink, ok := payload.(*models.PaymentLink)
	if !ok {
		return errors.New("invalid payload type, expected *models.PaymentLink")
	}
	logger := h.Service.Log(ctx).WithField("paymentLinkId", paymentLink.ID)
	logger.Info("Processing create.payment_link event")

	requestBody, err := h.prepareRequest(paymentLink)
	if err != nil {
		logger.WithError(err).Error("failed to prepare request")
		return h.handleError(ctx, paymentLink.ID, fmt.Errorf("prepare request: %w", err))
	}

	token, err := h.Client.GenerateBearerToken()
	if err != nil {
		logger.WithError(err).Error("failed to generate bearer token")
		return h.handleError(ctx, paymentLink.ID, fmt.Errorf("generate bearer token: %w", err))
	}

	response, err := h.Client.CreatePaymentLink(requestBody, token.AccessToken)
	if err != nil || !response.Status {
		errorMsg := h.getErrorResponse(err, response)
		logger.WithError(err).Error("failed to create payment link")
		return h.handleError(ctx, paymentLink.ID, fmt.Errorf("create payment link: %v", errorMsg))
	}

	h.logResponse(response)

	if err := h.updateStatusSuccess(ctx, paymentLink.ID); err != nil {
		logger.WithError(err).Error("failed to update payment link status")
		return fmt.Errorf("update payment status: %w", err)
	}
	return nil
}

func (h *CreatePaymentLink) prepareRequest(paymentLink *models.PaymentLink) (models.PaymentLinkRequest, error) {
	var customers []models.PaymentLinkCustomer
	if err := json.Unmarshal(paymentLink.Customers, &customers); err != nil {
		return models.PaymentLinkRequest{}, fmt.Errorf("unmarshal customers: %w", err)
	}

	var notifications []string
	if len(paymentLink.Notifications) > 0 {
		if err := json.Unmarshal(paymentLink.Notifications, &notifications); err != nil {
			return models.PaymentLinkRequest{}, fmt.Errorf("unmarshal notifications: %w", err)
		}
	}

	return models.PaymentLinkRequest{
		Customers: customers,
		PaymentLink: models.PaymentLinkDetails{
			ExpiryDate:      paymentLink.ExpiryDate.Format(dateFormat),
			SaleDate:        paymentLink.SaleDate.Format(dateFormat),
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
		},
		Notifications: notifications,
	}, nil
}

func (h *CreatePaymentLink) getErrorResponse(err error, response *models.PaymentLinkResponse) string {
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("API call failed with status: %v, message: %s", response.Status, response.Message)
}

func (h *CreatePaymentLink) logResponse(response *models.PaymentLinkResponse) {
	logger := h.Service.Log(context.Background())
	dataJSON, err := json.Marshal(response.Data)
	if err != nil {
		logger.WithError(err).Error("failed to marshal payment link data")
	} else {
		logger.WithField("paymentLinkData", string(dataJSON)).Info("Payment link data")
	}
	logger.WithField("response", response).Info("Payment link creation response received")
}

func (h *CreatePaymentLink) handleError(ctx context.Context, id string, err error) error {
	logger := h.Service.Log(ctx)
	statusUpdateRequest := &commonv1.StatusUpdateRequest{
		Id:     id,
		State:  statusActive,
		Status: statusFailed,
		Extras: map[string]string{
			"update_type": updateTypePaymentLink,
			"error":       err.Error(),
		},
	}

	if _, updateErr := h.PaymentClient.StatusUpdate(ctx, statusUpdateRequest); updateErr != nil {
		logger.WithError(updateErr).Error("failed to update payment link status")
		return fmt.Errorf("update payment status: %w", updateErr)
	}
	return err
}

func (h *CreatePaymentLink) updateStatusSuccess(ctx context.Context, id string) error {
	statusUpdateRequest := &commonv1.StatusUpdateRequest{
		Id:     id,
		State:  statusActive,
		Status: statusSuccessful,
		Extras: map[string]string{
			"update_type": updateTypePaymentLink,
			"message":     "Payment link successfully generated",
		},
	}

	_, err := h.PaymentClient.StatusUpdate(ctx, statusUpdateRequest)
	if err != nil {
		return fmt.Errorf("payment client status update: %w", err)
	}
	return nil
}
