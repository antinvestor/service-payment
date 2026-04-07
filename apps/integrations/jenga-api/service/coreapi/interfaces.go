package coreapi

import (
	"context"

	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
)

// Credentials holds per-request Jenga API credentials for multi-tenant support.
type Credentials struct {
	MerchantCode   string
	ConsumerSecret string
	APIKey         string
	Environment    string // Base URL e.g. https://uat.finserve.africa
	PrivateKeyPath string
}

//nolint:revive // JengaApiClient follows original API naming convention
type JengaApiClient interface { //nolint:staticcheck // API interface name
	GenerateBearerToken(ctx context.Context, creds *Credentials) (*BearerTokenResponse, error)
	InitiateSTKUSSD(ctx context.Context, creds *Credentials, request models.STKUSSDRequest, accessToken string) (*models.STKUSSDResponse, error)
	CreatePaymentLink(ctx context.Context, creds *Credentials, request models.PaymentLinkRequest, accessToken string) (*models.PaymentLinkResponse, error)
	InitiateTillsPay(ctx context.Context, creds *Credentials, request models.TillsPayRequest, accessToken string) (*models.TillsPayResponse, error)
}
