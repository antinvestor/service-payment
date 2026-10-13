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

package events_test

import (
	"context"
	"errors"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/stretchr/testify/require"
)

// Compile-time check: stub implements the client used by PaymentStatusUpdate.
var _ paymentv1connect.PaymentServiceClient = (*stubPaymentClient)(nil)

type stubPaymentClient struct {
	statusUpdateErr error
	calls           int
}

func (s *stubPaymentClient) StatusUpdate(
	_ context.Context,
	_ *commonv1.StatusUpdateRequest,
) (*commonv1.StatusUpdateResponse, error) {
	s.calls++
	if s.statusUpdateErr != nil {
		return nil, s.statusUpdateErr
	}
	return &commonv1.StatusUpdateResponse{
		Data: &commonv1.StatusResponse{Id: "prompt-1", Status: commonv1.STATUS_SUCCESSFUL},
	}, nil
}

func (s *stubPaymentClient) Send(context.Context, *paymentv1.SendRequest) (*paymentv1.SendResponse, error) {
	panic("unexpected")
}
func (s *stubPaymentClient) Receive(context.Context, *paymentv1.ReceiveRequest) (*paymentv1.ReceiveResponse, error) {
	panic("unexpected")
}
func (s *stubPaymentClient) InitiatePrompt(context.Context, *paymentv1.InitiatePromptRequest) (*paymentv1.InitiatePromptResponse, error) {
	panic("unexpected")
}
func (s *stubPaymentClient) CreatePaymentLink(context.Context, *paymentv1.CreatePaymentLinkRequest) (*paymentv1.CreatePaymentLinkResponse, error) {
	panic("unexpected")
}
func (s *stubPaymentClient) Status(context.Context, *commonv1.StatusRequest) (*commonv1.StatusResponse, error) {
	panic("unexpected")
}
func (s *stubPaymentClient) Release(context.Context, *paymentv1.ReleaseRequest) (*paymentv1.ReleaseResponse, error) {
	panic("unexpected")
}
func (s *stubPaymentClient) Search(context.Context, *commonv1.SearchRequest) (paymentv1connect.PaymentServiceSearchClientStream, error) {
	panic("unexpected")
}
func (s *stubPaymentClient) Reconcile(context.Context, *paymentv1.ReconcileRequest) (*paymentv1.ReconcileResponse, error) {
	panic("unexpected")
}

func TestPaymentStatusUpdate_PropagatesStatusUpdateError(t *testing.T) {
	t.Parallel()
	cli := &stubPaymentClient{statusUpdateErr: errors.New("permission_denied: payment_status_update")}
	h := events.NewPaymentStatusUpdate(context.Background(), cli)

	err := h.Execute(context.Background(), &commonv1.StatusUpdateRequest{
		Id:     "prompt-1",
		Status: commonv1.STATUS_SUCCESSFUL,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "permission_denied")
	require.Equal(t, 1, cli.calls)
}

func TestPaymentStatusUpdate_Succeeds(t *testing.T) {
	t.Parallel()
	cli := &stubPaymentClient{}
	h := events.NewPaymentStatusUpdate(context.Background(), cli)

	err := h.Execute(context.Background(), &commonv1.StatusUpdateRequest{
		Id:     "prompt-1",
		Status: commonv1.STATUS_SUCCESSFUL,
	})
	require.NoError(t, err)
	require.Equal(t, 1, cli.calls)
}
