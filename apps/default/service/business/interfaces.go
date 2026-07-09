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

package business

import (
	"context"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"github.com/pitabwire/frame/v2/workerpool"
)

type PaymentBusiness interface {
	Send(ctx context.Context, payment *paymentv1.Payment) (*commonv1.StatusResponse, error)
	Receive(ctx context.Context, payment *paymentv1.Payment) (*commonv1.StatusResponse, error)
	Status(ctx context.Context, status *commonv1.StatusRequest) (*commonv1.StatusResponse, error)
	StatusUpdate(ctx context.Context, req *commonv1.StatusUpdateRequest) (*commonv1.StatusResponse, error)
	Release(ctx context.Context, status *paymentv1.ReleaseRequest) (*commonv1.StatusResponse, error)
	Search(ctx context.Context, query *commonv1.SearchRequest) (workerpool.JobResultPipe[[]*paymentv1.Payment], error)
	InitiatePrompt(ctx context.Context, req *paymentv1.InitiatePromptRequest) (*commonv1.StatusResponse, error)
	CreatePaymentLink(ctx context.Context, req *paymentv1.CreatePaymentLinkRequest) (*commonv1.StatusResponse, error)
	Reconcile(ctx context.Context, msg *paymentv1.ReconcileRequest) (*paymentv1.ReconcileResponse, error)
}
