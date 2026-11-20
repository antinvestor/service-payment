// Copyright 2023-2026 Ant Investor Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package coreapi

import (
	"context"
	"strings"

	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/service/models"
)

// Jenga API hosts per environment.
const (
	EnvironmentProduction = "production"
	EnvironmentSandbox    = "sandbox"

	baseURLProduction = "https://api.finserve.africa"
	baseURLSandbox    = "https://uat.finserve.africa"
)

// Credentials holds per-request Jenga API credentials for multi-tenant support.
type Credentials struct {
	MerchantCode   string
	ConsumerSecret string
	APIKey         string
	// Environment is sandbox or production; it selects the API host when
	// BaseURL is empty.
	Environment string
	// BaseURL overrides the environment-derived API host (also used in tests).
	BaseURL string
	// PrivateKeyPEM is the RSA signing key. When empty the key is read from
	// PrivateKeyPath.
	PrivateKeyPEM  string
	PrivateKeyPath string
}

// ResolveBaseURL returns the Jenga API host for these credentials.
func (c *Credentials) ResolveBaseURL() string {
	if u := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/"); u != "" {
		return u
	}
	if strings.EqualFold(strings.TrimSpace(c.Environment), EnvironmentProduction) {
		return baseURLProduction
	}
	return baseURLSandbox
}

//nolint:revive // JengaApiClient follows original API naming convention
type JengaApiClient interface { //nolint:staticcheck // API interface name
	GenerateBearerToken(ctx context.Context, creds *Credentials) (*BearerTokenResponse, error)
	InitiateSTKUSSD(
		ctx context.Context,
		creds *Credentials,
		request models.STKUSSDRequest,
		accessToken string,
	) (*models.STKUSSDResponse, error)
	CreatePaymentLink(
		ctx context.Context,
		creds *Credentials,
		request models.PaymentLinkRequest,
		accessToken string,
	) (*models.PaymentLinkResponse, error)
	InitiateTillsPay(
		ctx context.Context,
		creds *Credentials,
		request models.TillsPayRequest,
		accessToken string,
	) (*models.TillsPayResponse, error)
}
