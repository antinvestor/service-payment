package coreapi

import (
	"github.com/antinvestor/jenga-api/service/models"
	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of the ClientInterface.
type MockClient struct {
	mock.Mock
}

// GenerateBearerToken mocks the GenerateBearerToken method.
func (m *MockClient) GenerateBearerToken() (*BearerTokenResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*BearerTokenResponse), args.Error(1)
}

// InitiateSTKUSSD mocks the InitiateSTKUSSD method.
func (m *MockClient) InitiateSTKUSSD(request models.STKUSSDRequest, accessToken string) (*models.STKUSSDResponse, error) {
	args := m.Called(request, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.STKUSSDResponse), args.Error(1)
}

// InitiateAccountBalance mocks the InitiateAccountBalance method.
//
//nolint:revive // accountId follows API parameter naming convention
func (m *MockClient) InitiateAccountBalance(
	countryCode, accountId, accessToken string,
) (*models.BalanceResponse, error) {
	args := m.Called(countryCode, accountId, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BalanceResponse), args.Error(1)
}

// FetchBillers mocks the FetchBillers method.
func (m *MockClient) FetchBillers(token string) ([]models.Biller, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Biller), args.Error(1)
}

// CreatePaymentLink mocks the CreatePaymentLink method.
func (m *MockClient) CreatePaymentLink(request models.PaymentLinkRequest, accessToken string) (*models.PaymentLinkResponse, error) {
	args := m.Called(request, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PaymentLinkResponse), args.Error(1)
}

// InitiateTillsPay mocks the InitiateTillsPay method.
func (m *MockClient) InitiateTillsPay(request models.TillsPayRequest, accessToken string) (*models.TillsPayResponse, error) {
	args := m.Called(request, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TillsPayResponse), args.Error(1)
}
