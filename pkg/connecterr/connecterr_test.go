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

package connecterr_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"connectrpc.com/connect/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/antinvestor/service-payments/pkg/connecterr"
)

func TestFrom(t *testing.T) {
	t.Parallel()

	downstream := connect.NewError(connect.CodePermissionDenied, "denied by ledger")
	grpcErr := status.Error(codes.FailedPrecondition, "already settled")

	tests := []struct {
		name    string
		err     error
		code    connect.Code
		message string
	}{
		{"connect error kept", downstream, connect.CodePermissionDenied, "denied by ledger"},
		{"grpc status keeps code", grpcErr, connect.CodeFailedPrecondition, "already settled"},
		{"wrapped grpc status keeps code", fmt.Errorf("release: %w", grpcErr), connect.CodeFailedPrecondition, "already settled"},
		{"context canceled", context.Canceled, connect.CodeCanceled, "context canceled"},
		{"deadline", fmt.Errorf("query: %w", context.DeadlineExceeded), connect.CodeDeadlineExceeded, "deadline exceeded"},
		{"record not found", gorm.ErrRecordNotFound, connect.CodeNotFound, "record not found"},
		{"plain error", errors.New("boom"), connect.CodeInternal, "boom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := connecterr.From(tt.err)

			var connectErr *connect.Error
			require.ErrorAs(t, got, &connectErr)
			require.Equal(t, tt.code, connectErr.Code())
			require.Contains(t, connectErr.Message(), tt.message)
			require.ErrorIs(t, got, tt.err)
		})
	}
}

func TestFromWrappedConnectErrorKeepsDownstreamError(t *testing.T) {
	t.Parallel()

	downstream := connect.NewError(connect.CodePermissionDenied, "denied by ledger")
	got := connecterr.From(fmt.Errorf("call ledger: %w", downstream))
	require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(got))
	require.ErrorIs(t, got, downstream)
}

func TestFromNil(t *testing.T) {
	t.Parallel()
	require.NoError(t, connecterr.From(nil))
}
