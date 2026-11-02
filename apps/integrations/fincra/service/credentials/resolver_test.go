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
	"errors"
	"testing"

	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/config"
	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSettings struct {
	settingsv1connect.SettingsServiceClient
	value string
	err   error
	got   *settingsv1.Setting
}

func (f *fakeSettings) Get(_ context.Context, req *settingsv1.GetRequest) (*settingsv1.GetResponse, error) {
	f.got = req.GetKey()
	if f.err != nil {
		return nil, f.err
	}
	return &settingsv1.GetResponse{Data: &settingsv1.SettingObject{Value: f.value}}, nil
}

func baseConfig() *config.FincraConfig {
	return &config.FincraConfig{
		SecretKey: "sk-env", PublicKey: "pk-env", BusinessID: "biz-env", Environment: "sandbox",
		CollectionMode: "checkout", FeeBearer: "business",
		SettingsIntegrationName: "Fincra", SettingsIntegrationID: "Default",
	}
}

func TestDefaultUsesConfig(t *testing.T) {
	creds, err := credentials.NewResolver(nil, baseConfig()).Default()
	require.NoError(t, err)
	assert.Equal(t, "sk-env", creds.SecretKey)
	assert.Equal(t, "pk-env", creds.PublicKey)
	assert.Equal(t, "biz-env", creds.BusinessID)
}

func TestMissingCredentials(t *testing.T) {
	cfg := baseConfig()
	cfg.BusinessID = ""
	_, err := credentials.NewResolver(nil, cfg).Default()
	assert.ErrorIs(t, err, credentials.ErrMissingCredentials)
}

func TestHeadersOverrideConfig(t *testing.T) {
	creds, err := credentials.NewResolver(nil, baseConfig()).FromHeaders(context.Background(), map[string]string{
		config.HeaderSecretKey:      "sk-header",
		config.HeaderCollectionMode: "direct",
	})
	require.NoError(t, err)
	assert.Equal(t, "sk-header", creds.SecretKey)
	assert.Equal(t, "biz-env", creds.BusinessID, "unset headers fall back to config")
	assert.Equal(t, "direct", creds.ResolveCollectionMode())
}

func TestConnectionFromSettings(t *testing.T) {
	settings := &fakeSettings{value: `{"X-FINCRA_SECRET_KEY":"sk-conn","X-FINCRA_BUSINESS_ID":"biz-conn",` +
		`"X-FINCRA_WEBHOOK_SECRET":"wh-conn","X-FINCRA_ENVIRONMENT":"production"}`}
	creds, err := credentials.NewResolver(settings, baseConfig()).FromHeaders(context.Background(), map[string]string{
		config.HeaderConnectionCredentials: "acme",
		config.HeaderSecretKey:             "ignored-when-a-connection-is-named",
	})
	require.NoError(t, err)
	assert.Equal(t, "acme", settings.got.GetName())
	assert.Equal(t, "Fincra", settings.got.GetModule())
	assert.Equal(t, "sk-conn", creds.SecretKey)
	assert.Equal(t, "biz-conn", creds.BusinessID)
	assert.Equal(t, "wh-conn", creds.WebhookSecret)
	assert.True(t, creds.IsProduction())
	assert.Empty(t, creds.PublicKey, "a connection never borrows the service's keys")
	assert.Equal(t, "checkout", creds.ResolveCollectionMode(), "non-secret settings still default to config")
}

func TestConnectionWithoutKeysIsRefused(t *testing.T) {
	settings := &fakeSettings{value: `{"X-FINCRA_TENANT_ID":"t1"}`}
	_, err := credentials.NewResolver(settings, baseConfig()).FromConnection(context.Background(), "acme")
	assert.ErrorIs(t, err, credentials.ErrMissingCredentials)
}

func TestConnectionErrors(t *testing.T) {
	_, err := credentials.NewResolver(nil, baseConfig()).FromConnection(context.Background(), "acme")
	require.Error(t, err, "no settings client")

	_, err = credentials.NewResolver(&fakeSettings{err: errors.New("down")}, baseConfig()).
		FromConnection(context.Background(), "acme")
	require.Error(t, err)

	_, err = credentials.NewResolver(&fakeSettings{value: "not json"}, baseConfig()).
		FromConnection(context.Background(), "acme")
	require.Error(t, err)
}
