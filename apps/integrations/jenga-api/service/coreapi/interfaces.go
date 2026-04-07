package coreapi

import (
	"context"

	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
)

//nolint:revive // JengaApiClient follows original API naming convention
type JengaApiClient interface { //nolint:staticcheck // API interface name
	GenerateBearerToken(ctx context.Context) (*BearerTokenResponse, error)
	InitiateSTKUSSD(ctx context.Context, request models.STKUSSDRequest, accessToken string) (*models.STKUSSDResponse, error)
	CreatePaymentLink(ctx context.Context, request models.PaymentLinkRequest, accessToken string) (*models.PaymentLinkResponse, error)
	InitiateTillsPay(ctx context.Context, request models.TillsPayRequest, accessToken string) (*models.TillsPayResponse, error)
}
