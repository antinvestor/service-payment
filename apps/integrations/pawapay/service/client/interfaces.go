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

// PawapayClient defines the interface for pawaPay Merchant API v2 operations.
// See https://docs.pawapay.io/v2/api-reference for the full specification.
type PawapayClient interface {
	// InitiateDeposit requests a payment from a customer's mobile money wallet (collection).
	InitiateDeposit(ctx context.Context, creds *Credentials, req *DepositRequest) (*DepositInitiationResponse, error)

	// GetDeposit checks the status of a previously initiated deposit.
	GetDeposit(ctx context.Context, creds *Credentials, depositID string) (*DepositStatusResult, error)

	// ResendDepositCallback requests pawaPay to resend the final callback for a deposit.
	ResendDepositCallback(ctx context.Context, creds *Credentials, depositID string) (*ManualActionResponse, error)

	// InitiatePayout sends money to a customer's mobile money wallet (disbursement).
	InitiatePayout(ctx context.Context, creds *Credentials, req *PayoutRequest) (*PayoutInitiationResponse, error)

	// InitiateBulkPayouts sends up to multiple payouts in a single request.
	InitiateBulkPayouts(
		ctx context.Context,
		creds *Credentials,
		reqs []*PayoutRequest,
	) ([]*PayoutInitiationResponse, error)

	// GetPayout checks the status of a previously initiated payout.
	GetPayout(ctx context.Context, creds *Credentials, payoutID string) (*PayoutStatusResult, error)

	// ResendPayoutCallback requests pawaPay to resend the final callback for a payout.
	ResendPayoutCallback(ctx context.Context, creds *Credentials, payoutID string) (*ManualActionResponse, error)

	// CancelEnqueuedPayout fails a payout that is enqueued waiting for provider availability.
	CancelEnqueuedPayout(ctx context.Context, creds *Credentials, payoutID string) (*ManualActionResponse, error)

	// InitiateRefund refunds a completed deposit back to the customer.
	InitiateRefund(ctx context.Context, creds *Credentials, req *RefundRequest) (*RefundInitiationResponse, error)

	// GetRefund checks the status of a previously initiated refund.
	GetRefund(ctx context.Context, creds *Credentials, refundID string) (*RefundStatusResult, error)

	// ResendRefundCallback requests pawaPay to resend the final callback for a refund.
	ResendRefundCallback(ctx context.Context, creds *Credentials, refundID string) (*ManualActionResponse, error)

	// CancelEnqueuedRefund fails a refund that is enqueued waiting for provider availability.
	CancelEnqueuedRefund(ctx context.Context, creds *Credentials, refundID string) (*ManualActionResponse, error)

	// CreatePaymentPageSession creates a hosted Payment Page session for a deposit.
	CreatePaymentPageSession(
		ctx context.Context,
		creds *Credentials,
		req *PaymentPageRequest,
	) (*PaymentPageSession, error)

	// PredictProvider sanitises a phone number (MSISDN) and predicts its mobile money provider.
	PredictProvider(ctx context.Context, creds *Credentials, phoneNumber string) (*ProviderPrediction, error)

	// ActiveConfiguration returns the providers, currencies and limits configured for the account.
	// country and operationType are optional filters.
	ActiveConfiguration(
		ctx context.Context,
		creds *Credentials,
		country, operationType string,
	) (*ActiveConfiguration, error)

	// Availability returns the current operational status of providers.
	// country and operationType are optional filters.
	Availability(ctx context.Context, creds *Credentials, country, operationType string) ([]CountryAvailability, error)

	// WalletBalances returns the balances of the account's pawaPay wallets.
	// country is an optional filter.
	WalletBalances(ctx context.Context, creds *Credentials, country string) (*WalletBalances, error)
}
