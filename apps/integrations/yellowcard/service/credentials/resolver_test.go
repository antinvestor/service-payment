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

package credentials_test

import (
	"context"
	"testing"

	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/config"
	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfig() *config.YellowcardConfig {
	return &config.YellowcardConfig{
		APIKey:       "cfg-key",
		SecretKey:    "cfg-secret",
		Environment:  "sandbox",
		Country:      "UG",
		CustomerType: "retail",
	}
}

func TestFromHeaders_Defaults(t *testing.T) {
	creds, err := credentials.NewResolver(nil, testConfig()).FromHeaders(t.Context(), map[string]string{})
	require.NoError(t, err)
	assert.Equal(t, "cfg-key", creds.APIKey)
	assert.Equal(t, "cfg-secret", creds.SecretKey)
	assert.Equal(t, "UG", creds.Country)
	assert.Equal(t, "retail", creds.CustomerType)
}

func TestFromHeaders_Override(t *testing.T) {
	headers := map[string]string{
		config.HeaderAPIKey:       "hdr-key",
		config.HeaderSecretKey:    "hdr-secret",
		config.HeaderEnvironment:  "production",
		config.HeaderNetwork:      "MTN",
		config.HeaderCustomerType: "institution",
		config.HeaderBusinessID:   "biz-1",
	}
	creds, err := credentials.NewResolver(nil, testConfig()).FromHeaders(t.Context(), headers)
	require.NoError(t, err)
	assert.Equal(t, "hdr-key", creds.APIKey)
	assert.Equal(t, "hdr-secret", creds.SecretKey)
	assert.Equal(t, "production", creds.Environment)
	assert.Equal(t, "MTN", creds.Network)
	assert.Equal(t, "institution", creds.CustomerType)
	assert.Equal(t, "biz-1", creds.BusinessID)
	assert.Equal(t, "UG", creds.Country, "unset header falls back to config")
}

func TestFromHeaders_Missing(t *testing.T) {
	cfg := testConfig()
	cfg.SecretKey = ""
	_, err := credentials.NewResolver(nil, cfg).FromHeaders(t.Context(), nil)
	require.ErrorIs(t, err, credentials.ErrMissingCredentials)
}

func TestFromHeaders_ConnectionWithoutSettings(t *testing.T) {
	headers := map[string]string{config.HeaderConnectionCredentials: "tenant-a"}
	_, err := credentials.NewResolver(nil, testConfig()).FromHeaders(t.Context(), headers)
	require.Error(t, err)
}

func TestDefault(t *testing.T) {
	cfg := testConfig()
	cfg.CustomerType = ""
	creds, err := credentials.NewResolver(nil, cfg).Default()
	require.NoError(t, err)
	assert.Equal(t, "cfg-key", creds.APIKey)
	assert.Equal(t, "retail", creds.CustomerType, "customer type defaults to retail")
}

type fakeSettings struct {
	settingsv1connect.SettingsServiceClient
	value string
}

func (f *fakeSettings) Get(_ context.Context, _ *settingsv1.GetRequest) (*settingsv1.GetResponse, error) {
	return &settingsv1.GetResponse{Data: &settingsv1.SettingObject{Value: f.value}}, nil
}

// A connection is a separate account: it never borrows the service's keys.
func TestFromConnection_NeverInheritsServiceKeys(t *testing.T) {
	cfg := testConfig()
	cfg.WebhookSecret = "cfg-webhook"

	_, err := credentials.NewResolver(&fakeSettings{value: `{"X-YELLOWCARD_TENANT_ID":"t1"}`}, cfg).
		FromConnection(t.Context(), "acme")
	require.ErrorIs(t, err, credentials.ErrMissingCredentials)

	creds, err := credentials.NewResolver(&fakeSettings{
		value: `{"X-YELLOWCARD_API_KEY":"k","X-YELLOWCARD_SECRET_KEY":"s","X-YELLOWCARD_TENANT_ID":"t1"}`,
	}, cfg).FromConnection(t.Context(), "acme")
	require.NoError(t, err)
	assert.Equal(t, "s", creds.ResolveWebhookSecret(), "the connection's own secret verifies its webhooks")
	assert.Equal(t, "UG", creds.Country, "non-secret defaults still come from config")
	assert.Equal(t, "acme", creds.Connection)
	assert.Equal(t, "t1", creds.TenantID)
}
