package coreapi

import "github.com/antinvestor/jenga-api/service/models"

//nolint:revive // JengaApiClient follows original API naming convention
type JengaApiClient interface {
	GenerateBearerToken() (*BearerTokenResponse, error)
	InitiateSTKUSSD(request models.STKUSSDRequest, accessToken string) (*models.STKUSSDResponse, error)
	CreatePaymentLink(request models.PaymentLinkRequest, accessToken string) (*models.PaymentLinkResponse, error)
	InitiateTillsPay(request models.TillsPayRequest, accessToken string) (*models.TillsPayResponse, error)
}
