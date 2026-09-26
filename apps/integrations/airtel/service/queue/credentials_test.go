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

package queue

import (
	"testing"

	"github.com/antinvestor/service-payments/apps/integrations/airtel/config"
	"github.com/antinvestor/service-payments/apps/integrations/airtel/service/client"
	"github.com/stretchr/testify/assert"
)

func TestWithRequestBinding(t *testing.T) {
	extras := withRequestBinding(map[string]any{client.ExtraTransactionID: "tx-1"}, "5000", "UGX",
		map[string]string{config.HeaderConnectionCredentials: "conn-1"})
	assert.Equal(t, "tx-1", extras[client.ExtraTransactionID])
	assert.Equal(t, "5000", extras[client.ExtraRequestedAmount])
	assert.Equal(t, "UGX", extras[client.ExtraRequestedCurrency])
	assert.Equal(t, "conn-1", extras[client.ExtraCredentialsConnection])

	extras = withRequestBinding(map[string]any{}, "1", "UGX", map[string]string{})
	assert.NotContains(t, extras, client.ExtraCredentialsConnection)
}
