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

// MpesaClient defines the interface for M-Pesa Daraja API operations.
type MpesaClient interface {
	// STKPush initiates an STK Push (Lipa Na M-Pesa) request to the customer's phone
	STKPush(ctx context.Context, creds *MpesaCredentials, req *STKPushRequest) (*STKPushResponse, error)

	// B2CPayment initiates a Business-to-Customer payment (disbursement)
	B2CPayment(ctx context.Context, creds *MpesaCredentials, req *B2CRequest) (*B2CResponse, error)
}
