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

package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v2/v1/paymentv1connect"
	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/antinvestor/service-payments/apps/default/service/business"
	"github.com/antinvestor/service-payments/pkg/connecterr"
)

// statusUpdateRecorder is the server side of the webhook/callback path: every
// integration forwards provider callbacks to PaymentService.StatusUpdate.
type statusUpdateRecorder struct {
	paymentv1connect.UnimplementedPaymentServiceHandler

	received int
	err      error
}

func (s *statusUpdateRecorder) StatusUpdate(
	_ context.Context,
	req *commonv1.StatusUpdateRequest,
) (*commonv1.StatusUpdateResponse, error) {
	if s.err != nil {
		return nil, connecterr.From(s.err)
	}
	s.received = proto.Size(req)
	return &commonv1.StatusUpdateResponse{}, nil
}

func newPaymentTestClient(t *testing.T, impl paymentv1connect.PaymentServiceHandler) paymentv1connect.PaymentServiceClient {
	t.Helper()

	server := connect.NewServer()
	paymentv1connect.RegisterPaymentServiceHandler(server, impl)

	mux := http.NewServeMux()
	mountPaymentServer(mux, server)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return paymentv1connect.NewPaymentServiceClient(connect.NewClient(
		connecthttp.NewTransport(srv.Client(), srv.URL),
	))
}

// statusUpdateOfSize builds a StatusUpdateRequest whose extras carry a
// callback payload of roughly n bytes.
func statusUpdateOfSize(t *testing.T, n int) *commonv1.StatusUpdateRequest {
	t.Helper()

	extras, err := structpb.NewStruct(map[string]any{
		"provider_callback": strings.Repeat("x", n),
	})
	require.NoError(t, err)

	return &commonv1.StatusUpdateRequest{Id: "payment-1", Extras: extras}
}

func TestPaymentReadLimit_AcceptsCallbackAboveConnectDefault(t *testing.T) {
	t.Parallel()

	const connectV2DefaultReadMaxBytes = 4 << 20
	require.Greater(t, paymentMaxRequestMessageBytes, connectV2DefaultReadMaxBytes)

	rec := &statusUpdateRecorder{}
	cli := newPaymentTestClient(t, rec)

	// 5 MiB: rejected by connect v2's default, accepted by the explicit limit.
	_, err := cli.StatusUpdate(t.Context(), statusUpdateOfSize(t, 5<<20))
	require.NoError(t, err)
	require.Greater(t, rec.received, connectV2DefaultReadMaxBytes)
}

func TestPaymentReadLimit_RejectsCallbackAboveLimit(t *testing.T) {
	t.Parallel()

	rec := &statusUpdateRecorder{}
	cli := newPaymentTestClient(t, rec)

	_, err := cli.StatusUpdate(t.Context(), statusUpdateOfSize(t, paymentMaxRequestMessageBytes+1024))
	require.Error(t, err)
	require.Equal(t, connect.CodeResourceExhausted, connect.CodeOf(err))
	require.Zero(t, rec.received, "handler must not run for an oversized message")
}

func TestPaymentErrors_KeepCodeAndMessageOnTheWire(t *testing.T) {
	t.Parallel()

	// business errors are gRPC status errors; connect v2 would send them as a
	// bare Unknown without connecterr.
	rec := &statusUpdateRecorder{err: business.ErrPaymentDoesNotExist}
	cli := newPaymentTestClient(t, rec)

	_, err := cli.StatusUpdate(t.Context(), statusUpdateOfSize(t, 16))
	require.Error(t, err)
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	require.Contains(t, err.Error(), "Specified payment does not exist")
}
