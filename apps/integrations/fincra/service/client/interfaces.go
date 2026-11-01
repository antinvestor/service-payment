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

package client

import "context"

// FincraClient is the subset of the Fincra API the integration uses.
type FincraClient interface {
	// CreateCheckout creates a hosted checkout link.
	CreateCheckout(ctx context.Context, creds *Credentials, req *CheckoutRequest) (*CheckoutLink, error)
	// GetCheckout verifies a hosted checkout by our reference.
	GetCheckout(ctx context.Context, creds *Credentials, merchantReference string) (*CheckoutPayment, error)

	// InitiateCharge starts a direct charge.
	InitiateCharge(ctx context.Context, creds *Credentials, req *ChargeRequest) (*Charge, error)
	// AuthorizeCharge completes a pending direct charge with an OTP.
	AuthorizeCharge(ctx context.Context, creds *Credentials, chargeID string, req *AuthorizeRequest) (*Charge, error)
	// GetCharge verifies a direct charge by our reference.
	GetCharge(ctx context.Context, creds *Credentials, merchantReference string) (*Charge, error)

	// CreatePayout sends a payout.
	CreatePayout(ctx context.Context, creds *Credentials, req *PayoutRequest) (*Payout, error)
	// GetPayout fetches a payout by our reference.
	GetPayout(ctx context.Context, creds *Credentials, customerReference string) (*Payout, error)

	// ListProviders lists banks or mobile money operators for a corridor.
	ListProviders(ctx context.Context, creds *Credentials, country, currency, destination string) ([]Provider, error)
}
