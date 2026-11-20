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
	"errors"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"connectrpc.com/connect/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/antinvestor/service-payments/apps/default/service/metrics"
	"github.com/antinvestor/service-payments/apps/default/service/models"
)

// A prompt on a route without an integration queue is refused up front
// instead of being published to the default topic nobody consumes.
func TestInitiatePrompt_RefusesRouteWithoutQueue(t *testing.T) {
	q := &publishedPrompts{}
	pb := &paymentBusiness{
		qMan:            q,
		promptTopicName: "initiate.prompt",
		promptRepo:      &promptsByID{byID: map[string]*models.Prompt{}},
		obs:             metrics.NewMetrics(),
	}
	pb.WithPromptRoutes([]string{"Fincra", " mpesa ", ""})
	assert.ElementsMatch(t, []string{"fincra", "mpesa"}, pb.PromptRoutes())

	_, err := pb.InitiatePrompt(context.Background(), &paymentv1.InitiatePromptRequest{
		Route:  "stripe",
		Source: &commonv1.ContactLink{ContactId: "+254700000000"},
		Amount: &commonv1.Money{CurrencyCode: "KES", Units: 10},
	})
	require.Error(t, err)
	var connectErr *connect.Error
	require.True(t, errors.As(err, &connectErr))
	assert.Equal(t, connect.CodeFailedPrecondition, connectErr.Code())
	assert.ErrorIs(t, err, ErrPromptRouteUnavailable)
	assert.Empty(t, q.payloads, "nothing may be queued for an unserved route")
}
