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

// YellowcardClient defines the Yellow Card Payments API operations used by
// the integration. See https://docs.yellowcard.engineering/ for the spec.
// Credentials are supplied per call so one client serves many tenants.
type YellowcardClient interface {
	// SubmitReceive creates a receive (collection) request. With ForceAccept
	// the request is accepted immediately and processing starts.
	SubmitReceive(ctx context.Context, creds *Credentials, req *ReceiveRequest) (*Receive, error)
	// AcceptReceive accepts a quoted receive request for execution.
	AcceptReceive(ctx context.Context, creds *Credentials, id string) (*Receive, error)
	// DenyReceive rejects a quoted receive request.
	DenyReceive(ctx context.Context, creds *Credentials, id string) (*Receive, error)
	// CancelReceive cancels a receive request (Nigeria bank receives).
	CancelReceive(ctx context.Context, creds *Credentials, id string) (*Receive, error)
	// RefundReceive refunds a completed or cancelled receive (Nigeria bank receives).
	RefundReceive(ctx context.Context, creds *Credentials, id string) (*Receive, error)
	// GetReceive looks up a receive by its Yellow Card id.
	GetReceive(ctx context.Context, creds *Credentials, id string) (*Receive, error)
	// GetReceiveBySequenceID looks up a receive by the partner sequence id.
	GetReceiveBySequenceID(ctx context.Context, creds *Credentials, sequenceID string) (*Receive, error)

	// SubmitSend creates a send (disbursement) request.
	SubmitSend(ctx context.Context, creds *Credentials, req *SendRequest) (*Send, error)
	// AcceptSend accepts a quoted send request for execution.
	AcceptSend(ctx context.Context, creds *Credentials, id string) (*Send, error)
	// DenySend rejects a quoted send request.
	DenySend(ctx context.Context, creds *Credentials, id string) (*Send, error)
	// GetSend looks up a send by its Yellow Card id.
	GetSend(ctx context.Context, creds *Credentials, id string) (*Send, error)
	// GetSendBySequenceID looks up a send by the partner sequence id.
	GetSendBySequenceID(ctx context.Context, creds *Credentials, sequenceID string) (*Send, error)

	// GetChannels lists payment channels, optionally filtered by ISO alpha-2 country.
	GetChannels(ctx context.Context, creds *Credentials, country string) ([]Channel, error)
	// GetNetworks lists banks and mobile money networks, optionally filtered by country.
	GetNetworks(ctx context.Context, creds *Credentials, country string) ([]Network, error)
	// GetRates lists partner rates, optionally filtered by ISO 4217 currency.
	GetRates(ctx context.Context, creds *Credentials, currency string) ([]Rate, error)
}
