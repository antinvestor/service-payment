package coreapi

import (
	"context"

	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of the JengaApiClient interface.
type MockClient struct {
	mock.Mock
}

// GenerateBearerToken mocks the GenerateBearerToken method.
func (m *MockClient) GenerateBearerToken(ctx context.Context) (*BearerTokenResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	resp, ok := args.Get(0).(*BearerTokenResponse)
	if !ok {
		return nil, args.Error(1)
	}
	return resp, args.Error(1)
}

// InitiateSTKUSSD mocks the InitiateSTKUSSD method.
func (m *MockClient) InitiateSTKUSSD(
	ctx context.Context,
	request models.STKUSSDRequest,
	accessToken string,
) (*models.STKUSSDResponse, error) {
	args := m.Called(ctx, request, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	resp, ok := args.Get(0).(*models.STKUSSDResponse)
	if !ok {
		return nil, args.Error(1)
	}
	return resp, args.Error(1)
}

// CreatePaymentLink mocks the CreatePaymentLink method.
func (m *MockClient) CreatePaymentLink(
	ctx context.Context,
	request models.PaymentLinkRequest,
	accessToken string,
) (*models.PaymentLinkResponse, error) {
	args := m.Called(ctx, request, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	resp, ok := args.Get(0).(*models.PaymentLinkResponse)
	if !ok {
		return nil, args.Error(1)
	}
	return resp, args.Error(1)
}

// InitiateTillsPay mocks the InitiateTillsPay method.
func (m *MockClient) InitiateTillsPay(
	ctx context.Context,
	request models.TillsPayRequest,
	accessToken string,
) (*models.TillsPayResponse, error) {
	args := m.Called(ctx, request, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	resp, ok := args.Get(0).(*models.TillsPayResponse)
	if !ok {
		return nil, args.Error(1)
	}
	return resp, args.Error(1)
}
