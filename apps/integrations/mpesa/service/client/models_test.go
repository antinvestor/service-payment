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

package client_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/antinvestor/service-payments/apps/integrations/mpesa/service/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSTKQueryResponse_ResultCodeStringOrNumber(t *testing.T) {
	tests := []struct {
		name string
		body string
		want client.FlexString
	}{
		{name: "string", body: `{"ResultCode":"0"}`, want: "0"},
		{name: "number", body: `{"ResultCode":1032}`, want: "1032"},
		{name: "missing", body: `{"ResponseCode":"0"}`, want: ""},
		{name: "null", body: `{"ResultCode":null}`, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp client.STKQueryResponse
			require.NoError(t, json.Unmarshal([]byte(tt.body), &resp))
			assert.Equal(t, tt.want, resp.ResultCode)
		})
	}
}

func TestNewSTKQueryRequest(t *testing.T) {
	creds := &client.MpesaCredentials{Shortcode: "174379", Passkey: "pk"}
	now := time.Date(2026, 9, 26, 13, 4, 5, 0, time.UTC)

	req := client.NewSTKQueryRequest(creds, "ws_CO_1", now)

	assert.Equal(t, "174379", req.BusinessShortCode)
	assert.Equal(t, "20260926130405", req.Timestamp)
	assert.Equal(t, "ws_CO_1", req.CheckoutRequestID)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("174379pk20260926130405")), req.Password)

	raw, err := json.Marshal(req)
	require.NoError(t, err)
	assert.JSONEq(t,
		`{"BusinessShortCode":"174379","Password":"`+req.Password+`","Timestamp":"20260926130405","CheckoutRequestID":"ws_CO_1"}`,
		string(raw))
}
