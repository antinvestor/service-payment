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
	"context"
	"testing"
	"time"

	"github.com/antinvestor/service-payments/apps/integrations/yellowcard/service/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countingClient struct {
	client.YellowcardClient
	channelCalls int
	networkCalls int
}

func (c *countingClient) GetChannels(_ context.Context, _ *client.Credentials, _ string) ([]client.Channel, error) {
	c.channelCalls++
	return []client.Channel{{ID: "c1", Status: "active", ChannelType: "momo", RampType: "deposit", Currency: "UGX"}}, nil
}

func (c *countingClient) GetNetworks(_ context.Context, _ *client.Credentials, _ string) ([]client.Network, error) {
	c.networkCalls++
	return []client.Network{{ID: "n1", Status: "active"}}, nil
}

func TestCatalogCaches(t *testing.T) {
	creds := &client.Credentials{APIKey: "k"}
	fake := &countingClient{}
	cat := client.NewCatalog(fake, time.Minute)

	_, err := cat.Channels(t.Context(), creds, "UG")
	require.NoError(t, err)
	_, err = cat.Channels(t.Context(), creds, "ug")
	require.NoError(t, err)
	assert.Equal(t, 1, fake.channelCalls, "second call within TTL is served from cache")

	_, err = cat.Channels(t.Context(), creds, "NG")
	require.NoError(t, err)
	assert.Equal(t, 2, fake.channelCalls, "different country is a different entry")

	_, err = cat.Networks(t.Context(), creds, "UG")
	require.NoError(t, err)
	_, err = cat.Networks(t.Context(), creds, "UG")
	require.NoError(t, err)
	assert.Equal(t, 1, fake.networkCalls)

	// Zero TTL disables caching.
	fake = &countingClient{}
	cat = client.NewCatalog(fake, 0)
	_, _ = cat.Channels(t.Context(), creds, "UG")
	_, _ = cat.Channels(t.Context(), creds, "UG")
	assert.Equal(t, 2, fake.channelCalls)
}

func TestSelectChannel(t *testing.T) {
	channels := []client.Channel{
		{ID: "inactive", Status: "inactive", ChannelType: "momo", RampType: "deposit", Currency: "UGX"},
		{ID: "api-off", Status: "active", APIStatus: "inactive", ChannelType: "momo", RampType: "deposit", Currency: "UGX"},
		{ID: "withdraw", Status: "active", ChannelType: "momo", RampType: "withdraw", Currency: "UGX"},
		{ID: "bank", Status: "active", ChannelType: "bank", RampType: "deposit", Currency: "UGX"},
		{ID: "momo", Status: "active", APIStatus: "active", ChannelType: "momo", RampType: "deposit", Currency: "UGX"},
	}

	ch, ok := client.SelectChannel(channels, "UGX", "momo", "deposit")
	require.True(t, ok)
	assert.Equal(t, "momo", ch.ID)

	ch, ok = client.SelectChannel(channels, "", "bank", "deposit")
	require.True(t, ok)
	assert.Equal(t, "bank", ch.ID)

	_, ok = client.SelectChannel(channels, "KES", "momo", "deposit")
	assert.False(t, ok)

	assert.True(t, client.HasActiveChannel(channels, "momo", "withdraw"))
	assert.False(t, client.HasActiveChannel(channels, "bank", "withdraw"))
}

func TestResolveNetwork(t *testing.T) {
	networks := []client.Network{
		{ID: "n-bank", Code: "063", Name: "Diamond Bank", Status: "active", AccountNumberType: "bank", ChannelIDs: []string{"c-bank"}},
		{ID: "n-off", Code: "OFF", Name: "Old Telco", Status: "inactive", AccountNumberType: "momo"},
		{ID: "n-mtn", Code: "MTN", Name: "MTN Uganda", Status: "active", AccountNumberType: "momo", ChannelIDs: []string{"c-momo"}},
		{ID: "n-airtel", Code: "AIRTEL", Name: "Airtel Uganda", Status: "active", AccountNumberType: "momo", ChannelIDs: []string{"c-momo"}},
	}
	momo := &client.Channel{ID: "c-momo"}

	tests := []struct {
		name    string
		hint    string
		channel *client.Channel
		acct    string
		wantID  string
		wantOK  bool
	}{
		{"hint by id", "n-airtel", momo, "momo", "n-airtel", true},
		{"hint by code fold", "mtn", momo, "momo", "n-mtn", true},
		{"hint by name", "Airtel Uganda", momo, "momo", "n-airtel", true},
		{"hint by partial name", "airtel", momo, "momo", "n-airtel", true},
		{"no hint first active momo on channel", "", momo, "momo", "n-mtn", true},
		{"no hint bank", "", &client.Channel{ID: "c-bank"}, "bank", "n-bank", true},
		{"no hint no channel", "", nil, "bank", "n-bank", true},
		{"nothing matches", "", &client.Channel{ID: "c-none"}, "momo", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, ok := client.ResolveNetwork(networks, tt.hint, tt.channel, tt.acct)
			assert.Equal(t, tt.wantOK, ok)
			if ok {
				assert.Equal(t, tt.wantID, n.ID)
			}
		})
	}
}
