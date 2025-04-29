package coreapi

import "github.com/antinvestor/jenga-api/service/models"

type JengaApiClient interface {
	GenerateBearerToken() (*BearerTokenResponse, error)
	InitiateSTKUSSD(request models.STKUSSDRequest, accessToken string) (*models.STKUSSDResponse, error)
	InitiateAccountBalance(countryCode string, accountId string, accessToken string) (*models.BalanceResponse, error)
}