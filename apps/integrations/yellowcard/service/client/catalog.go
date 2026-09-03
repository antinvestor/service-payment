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
	"strings"
	"sync"
	"time"
)

// Catalog caches Yellow Card channels and networks per credential and
// country so every payment does not repeat the catalog calls.
type Catalog struct {
	cli YellowcardClient
	ttl time.Duration
	now func() time.Time

	mu       sync.Mutex
	channels map[string]catalogEntry[[]Channel]
	networks map[string]catalogEntry[[]Network]
}

type catalogEntry[T any] struct {
	value   T
	expires time.Time
}

// NewCatalog creates a catalog cache. A non-positive ttl disables caching.
func NewCatalog(cli YellowcardClient, ttl time.Duration) *Catalog {
	return &Catalog{
		cli:      cli,
		ttl:      ttl,
		now:      time.Now,
		channels: map[string]catalogEntry[[]Channel]{},
		networks: map[string]catalogEntry[[]Network]{},
	}
}

func catalogKey(creds *Credentials, country string) string {
	return creds.APIKey + "|" + creds.ResolveBaseURL() + "|" + strings.ToUpper(country)
}

// Channels returns the channels for a country, cached for the TTL.
func (c *Catalog) Channels(ctx context.Context, creds *Credentials, country string) ([]Channel, error) {
	key := catalogKey(creds, country)
	c.mu.Lock()
	entry, ok := c.channels[key]
	c.mu.Unlock()
	if ok && c.now().Before(entry.expires) {
		return entry.value, nil
	}

	channels, err := c.cli.GetChannels(ctx, creds, country)
	if err != nil {
		return nil, err
	}
	if c.ttl > 0 {
		c.mu.Lock()
		c.channels[key] = catalogEntry[[]Channel]{value: channels, expires: c.now().Add(c.ttl)}
		c.mu.Unlock()
	}
	return channels, nil
}

// Networks returns the networks for a country, cached for the TTL.
func (c *Catalog) Networks(ctx context.Context, creds *Credentials, country string) ([]Network, error) {
	key := catalogKey(creds, country)
	c.mu.Lock()
	entry, ok := c.networks[key]
	c.mu.Unlock()
	if ok && c.now().Before(entry.expires) {
		return entry.value, nil
	}

	networks, err := c.cli.GetNetworks(ctx, creds, country)
	if err != nil {
		return nil, err
	}
	if c.ttl > 0 {
		c.mu.Lock()
		c.networks[key] = catalogEntry[[]Network]{value: networks, expires: c.now().Add(c.ttl)}
		c.mu.Unlock()
	}
	return networks, nil
}

// channelUsable reports whether a channel is active for API use.
func channelUsable(ch Channel) bool {
	if !strings.EqualFold(ch.Status, ChannelStatusActive) {
		return false
	}
	return ch.APIStatus == "" || strings.EqualFold(ch.APIStatus, ChannelStatusActive)
}

// SelectChannel picks the first usable channel matching the currency (when
// given), channel type and ramp type.
func SelectChannel(channels []Channel, currency, channelType, rampType string) (*Channel, bool) {
	for i := range channels {
		ch := channels[i]
		if !channelUsable(ch) {
			continue
		}
		if channelType != "" && !strings.EqualFold(ch.ChannelType, channelType) {
			continue
		}
		if rampType != "" && !strings.EqualFold(ch.RampType, rampType) {
			continue
		}
		if currency != "" && !strings.EqualFold(ch.Currency, currency) {
			continue
		}
		return &ch, true
	}
	return nil, false
}

// HasActiveChannel reports whether any usable channel exists for the type and ramp.
func HasActiveChannel(channels []Channel, channelType, rampType string) bool {
	_, ok := SelectChannel(channels, "", channelType, rampType)
	return ok
}

// ResolveNetwork picks the network to route through. A hint matching a
// network id, code or name wins. Otherwise active networks of the wanted
// account type (and linked to the channel, when given) are considered; a
// single candidate is returned, else the first candidate.
func ResolveNetwork(networks []Network, hint string, channel *Channel, accountType string) (*Network, bool) {
	hint = strings.TrimSpace(hint)
	if hint != "" {
		for i := range networks {
			n := networks[i]
			if strings.EqualFold(n.ID, hint) || strings.EqualFold(n.Code, hint) || strings.EqualFold(n.Name, hint) {
				return &n, true
			}
		}
		for i := range networks {
			n := networks[i]
			if strings.Contains(strings.ToLower(n.Name), strings.ToLower(hint)) {
				return &n, true
			}
		}
	}

	var candidates []Network
	for _, n := range networks {
		if !strings.EqualFold(n.Status, ChannelStatusActive) {
			continue
		}
		if accountType != "" && n.AccountNumberType != "" && !strings.EqualFold(n.AccountNumberType, accountType) {
			continue
		}
		if channel != nil && len(n.ChannelIDs) > 0 && !containsFold(n.ChannelIDs, channel.ID) {
			continue
		}
		candidates = append(candidates, n)
	}
	if len(candidates) == 0 {
		return nil, false
	}
	return &candidates[0], true
}

func containsFold(list []string, v string) bool {
	for _, s := range list {
		if strings.EqualFold(s, v) {
			return true
		}
	}
	return false
}
