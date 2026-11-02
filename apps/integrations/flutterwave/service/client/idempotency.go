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

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type idempotencyKeyCtx struct{}

// WithIdempotencyKey names the operation a write performs. Flutterwave
// answers a repeated X-Idempotency-Key with the original response, so a
// retried or redelivered write built from the same key is performed once.
// Keys must be derived from our own ids (a prompt, a payment), never random.
func WithIdempotencyKey(ctx context.Context, parts ...string) context.Context {
	return context.WithValue(ctx, idempotencyKeyCtx{}, IdempotencyKey(parts...))
}

// IdempotencyKey joins parts into a key; long keys are hashed so they stay
// within Flutterwave's header limits.
func IdempotencyKey(parts ...string) string {
	key := strings.Join(parts, ":")
	if len(key) <= 64 { //nolint:mnd // header length bound
		return key
	}
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func idempotencyKeyFrom(ctx context.Context) string {
	if v, ok := ctx.Value(idempotencyKeyCtx{}).(string); ok {
		return v
	}
	return ""
}
