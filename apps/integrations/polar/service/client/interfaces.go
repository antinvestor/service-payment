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

// PolarClient defines the interface for Polar.sh API operations.
type PolarClient interface {
	// CreateCheckout creates a Polar checkout session for requesting money
	CreateCheckout(ctx context.Context, creds *PolarCredentials, req *CheckoutRequest) (*CheckoutResponse, error)

	// VerifyWebhookSignature verifies a Polar webhook signature
	VerifyWebhookSignature(payload []byte, headers map[string]string, webhookSecret string) (*WebhookEvent, error)
}
