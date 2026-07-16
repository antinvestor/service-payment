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

// FlutterwaveClient is the Flutterwave v4 adapter.
// Docs: https://developer.flutterwave.com/docs/getting-started
type FlutterwaveClient interface {
	// CreateOrchestratorCharge creates customer + payment method + charge in one call.
	// Preferred for MoMo / bank transfer / opay collections.
	CreateOrchestratorCharge(ctx context.Context, creds *Credentials, req *OrchestratorChargeRequest) (*Charge, error)

	// GetCharge verifies a charge by id (GET /charges/{id}).
	GetCharge(ctx context.Context, creds *Credentials, chargeID string) (*Charge, error)

	// CreateTransferRecipient creates a payout recipient (POST /transfers/recipients).
	CreateTransferRecipient(ctx context.Context, creds *Credentials, req *TransferRecipientRequest) (recipientID string, err error)

	// CreateTransfer initiates an instant transfer (POST /transfers).
	CreateTransfer(ctx context.Context, creds *Credentials, req map[string]any) (*Transfer, error)

	// CreateDirectTransfer initiates orchestrator payout (POST /direct-transfers).
	CreateDirectTransfer(ctx context.Context, creds *Credentials, req *DirectTransferRequest) (*Transfer, error)

	// GetTransfer fetches transfer status (GET /transfers/{id}).
	GetTransfer(ctx context.Context, creds *Credentials, transferID string) (*Transfer, error)

	// VerifyWebhookSignature validates flutterwave-signature HMAC-SHA256(body, secret).
	VerifyWebhookSignature(rawBody []byte, signatureHeader, secretHash string) bool
}
