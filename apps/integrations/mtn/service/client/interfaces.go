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

// MtnClient defines the interface for MTN MoMo API operations.
type MtnClient interface {
	// RequestToPay initiates a request-to-pay (collection) from a customer
	RequestToPay(ctx context.Context, creds *MtnCredentials, req *RequestToPayRequest) error

	// GetRequestToPayStatus checks the status of a request-to-pay
	GetRequestToPayStatus(ctx context.Context, creds *MtnCredentials, referenceID string) (*RequestToPayStatus, error)

	// Transfer initiates a disbursement transfer to a customer
	Transfer(ctx context.Context, creds *MtnCredentials, req *TransferRequest) error

	// GetTransferStatus checks the status of a transfer
	GetTransferStatus(ctx context.Context, creds *MtnCredentials, referenceID string) (*TransferStatus, error)
}
