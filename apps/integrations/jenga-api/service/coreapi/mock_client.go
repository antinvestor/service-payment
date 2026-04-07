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

func (m *MockClient) GenerateBearerToken(ctx context.Context, creds *Credentials) (*BearerTokenResponse, error) {
	args := m.Called(ctx, creds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	resp, ok := args.Get(0).(*BearerTokenResponse)
	if !ok {
		return nil, args.Error(1)
	}
	return resp, args.Error(1)
}

func (m *MockClient) InitiateSTKUSSD(
	ctx context.Context,
	creds *Credentials,
	request models.STKUSSDRequest,
	accessToken string,
) (*models.STKUSSDResponse, error) {
	args := m.Called(ctx, creds, request, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	resp, ok := args.Get(0).(*models.STKUSSDResponse)
	if !ok {
		return nil, args.Error(1)
	}
	return resp, args.Error(1)
}

func (m *MockClient) CreatePaymentLink(
	ctx context.Context,
	creds *Credentials,
	request models.PaymentLinkRequest,
	accessToken string,
) (*models.PaymentLinkResponse, error) {
	args := m.Called(ctx, creds, request, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	resp, ok := args.Get(0).(*models.PaymentLinkResponse)
	if !ok {
		return nil, args.Error(1)
	}
	return resp, args.Error(1)
}

func (m *MockClient) InitiateTillsPay(
	ctx context.Context,
	creds *Credentials,
	request models.TillsPayRequest,
	accessToken string,
) (*models.TillsPayResponse, error) {
	args := m.Called(ctx, creds, request, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	resp, ok := args.Get(0).(*models.TillsPayResponse)
	if !ok {
		return nil, args.Error(1)
	}
	return resp, args.Error(1)
}
