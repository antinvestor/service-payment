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
	"context"
	"testing"

	"buf.build/gen/go/antinvestor/settingz/connectrpc/go/v2/settings/v1/settingsv1connect"
	settingsv1 "buf.build/gen/go/antinvestor/settingz/protocolbuffers/go/settings/v1"
	"github.com/antinvestor/service-payments/apps/integrations/jenga-api/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type settingsWith struct {
	settingsv1connect.SettingsServiceClient
	value string
}

func (s settingsWith) Get(_ context.Context, _ *settingsv1.GetRequest) (*settingsv1.GetResponse, error) {
	return &settingsv1.GetResponse{Data: &settingsv1.SettingObject{Value: s.value}}, nil
}

func tenantSetting(extra string) string {
	return `{"` + config.HeaderMerchantCode + `":"MC-T","` + config.HeaderConsumerSecret + `":"cs-t",` +
		`"` + config.HeaderAPIKey + `":"ak-t"` + extra + `}`
}

func TestExtractCredentials_SigningKeyAndHostRestrictions(t *testing.T) {
	serviceCfg := func() *config.JengaConfig {
		return &config.JengaConfig{
			MerchantCode: "MC", ConsumerSecret: "cs", ApiKey: "ak", JengaPrivateKeyPEM: "service-pem",
			Environment: "sandbox", JengaCallbackURL: "https://hooks.example.com",
		}
	}
	conn := map[string]string{config.HeaderConnectionCredentials: "tenant-1"}

	tests := []struct {
		name    string
		cfg     func() *config.JengaConfig
		setting string
		headers map[string]string
		wantErr error
		wantURL string
		wantPEM string
	}{
		{
			name: "service credentials use the service key", cfg: serviceCfg,
			wantPEM: "service-pem", wantURL: "https://uat.finserve.africa",
		},
		{
			name: "tenant connection without its own key is refused", cfg: serviceCfg,
			setting: tenantSetting(""), headers: conn, wantErr: errTenantKeyRequired,
		},
		{
			name: "tenant connection with its own key and a known host", cfg: serviceCfg,
			setting: tenantSetting(`,"` + config.HeaderPrivateKeyPEM + `":"tenant-pem","` + config.HeaderBaseURL + `":"https://api.finserve.africa"`),
			headers: conn, wantPEM: "tenant-pem", wantURL: "https://api.finserve.africa",
		},
		{
			name: "tenant connection pointing the signer at another host is refused", cfg: serviceCfg,
			setting: tenantSetting(`,"` + config.HeaderPrivateKeyPEM + `":"tenant-pem","` + config.HeaderBaseURL + `":"https://evil.example.com"`),
			headers: conn, wantErr: errBaseURLNotAllowed,
		},
		{
			name: "known host over plain http is refused", cfg: serviceCfg,
			setting: tenantSetting(`,"` + config.HeaderPrivateKeyPEM + `":"tenant-pem","` + config.HeaderBaseURL + `":"http://uat.finserve.africa"`),
			headers: conn, wantErr: errBaseURLNotAllowed,
		},
		{
			name: "allowlist override admits a test host",
			cfg: func() *config.JengaConfig {
				c := serviceCfg()
				c.BaseURL, c.BaseURLAllowlist = "http://127.0.0.1:9999", "127.0.0.1:9999"
				return c
			},
			wantPEM: "service-pem", wantURL: "http://127.0.0.1:9999",
		},
		{
			name: "service base URL outside the allowlist is refused",
			cfg: func() *config.JengaConfig {
				c := serviceCfg()
				c.BaseURL = "https://evil.example.com"
				return c
			},
			wantErr: errBaseURLNotAllowed,
		},
		{
			name: "header overrides of merchant credentials need their own key", cfg: serviceCfg,
			headers: map[string]string{config.HeaderMerchantCode: "OTHER"}, wantErr: errTenantKeyRequired,
		},
		{
			name: "a header key path is never accepted as the tenant's own key", cfg: serviceCfg,
			headers: map[string]string{config.HeaderMerchantCode: "OTHER", config.HeaderPrivateKeyPath: "/etc/jenga/service.key"},
			wantErr: errTenantKeyRequired,
		},
		{
			name: "a tenant setting that only names a key path is refused", cfg: serviceCfg,
			setting: tenantSetting(`,"` + config.HeaderPrivateKeyPath + `":"/etc/jenga/service.key"`),
			headers: conn, wantErr: errTenantKeyRequired,
		},
		{
			name: "an overridden connection never falls back to the service key file",
			cfg: func() *config.JengaConfig {
				c := serviceCfg()
				c.JengaPrivateKeyPEM, c.JengaPrivateKey = "", "/etc/jenga/service.key"
				return c
			},
			headers: map[string]string{config.HeaderMerchantCode: "OTHER", config.HeaderPrivateKeyPEM: "tenant-pem"},
			wantPEM: "tenant-pem", wantURL: "https://uat.finserve.africa",
		},
		{
			name: "header override of the host needs its own key and an allowed host", cfg: serviceCfg,
			headers: map[string]string{config.HeaderBaseURL: "https://evil.example.com", config.HeaderPrivateKeyPEM: "p"},
			wantErr: errBaseURLNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &credentialResolver{cfg: tt.cfg()}
			if tt.setting != "" {
				r.settingsCli = settingsWith{value: tt.setting}
			}
			creds, err := r.extractCredentials(t.Context(), tt.headers)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, creds)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPEM, creds.PrivateKeyPEM)
			if tt.wantPEM != "service-pem" {
				assert.Empty(t, creds.PrivateKeyPath, "a tenant or header connection never uses the service key file")
			}
			assert.Equal(t, tt.wantURL, toAPICreds(creds).ResolveBaseURL())
		})
	}
}
