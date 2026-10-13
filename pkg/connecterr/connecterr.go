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

// Package connecterr converts handler errors into connect errors.
//
// connect-go v2 only sends a *connect.Error that the handler built. Any other
// error, including a gRPC status error, reaches the client as a bare
// CodeUnknown with no message. Handlers pass their errors through From so the
// client keeps the code and message it received under connect-go v1.
package connecterr

import (
	"context"
	"errors"

	"connectrpc.com/connect/v2"
	"github.com/pitabwire/frame/v3/data"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// From converts err into a *connect.Error, keeping the code and message:
//   - a *connect.Error anywhere in the chain is returned as is;
//   - a gRPC status error keeps its code and message (the codes are numerically
//     identical in gRPC and connect);
//   - context cancellation and deadline errors map to Canceled and DeadlineExceeded;
//   - anything else goes through frame's data.ErrorConvertToAPI.
//
// The original error is attached with WithCause, so errors.Is and errors.As
// still match it on the server.
func From(err error) error {
	if err == nil {
		return nil
	}

	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return connectErr
	}

	if st, ok := status.FromError(err); ok && st.Code() != codes.OK && st.Code() != codes.Unknown {
		//nolint:gosec // gRPC and connect codes share the same numeric values (1-16).
		return connect.NewError(connect.Code(st.Code()), st.Message()).WithCause(err)
	}

	switch {
	case errors.Is(err, context.Canceled):
		return connect.NewError(connect.CodeCanceled, err.Error()).WithCause(err)
	case errors.Is(err, context.DeadlineExceeded):
		return connect.NewError(connect.CodeDeadlineExceeded, err.Error()).WithCause(err)
	}

	return data.ErrorConvertToAPI(err)
}
