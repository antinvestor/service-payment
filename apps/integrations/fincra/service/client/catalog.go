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

// Catalog caches Fincra's provider lists (GET /core/banks) per business,
// corridor and payment destination.
type Catalog struct {
	cli FincraClient
	ttl time.Duration
	now func() time.Time

	mu      sync.Mutex
	entries map[string]catalogEntry
}

type catalogEntry struct {
	providers []Provider
	expires   time.Time
}

// NewCatalog creates a provider catalog; a non-positive ttl disables caching.
func NewCatalog(cli FincraClient, ttl time.Duration) *Catalog {
	return &Catalog{cli: cli, ttl: ttl, now: time.Now, entries: map[string]catalogEntry{}}
}

// Providers returns the enabled providers for a corridor and destination.
func (c *Catalog) Providers(
	ctx context.Context,
	creds *Credentials,
	country, currency, destination string,
) ([]Provider, error) {
	key := strings.Join([]string{creds.ResolveBaseURL(), creds.BusinessID, country, currency, destination}, "|")
	if c.ttl > 0 {
		c.mu.Lock()
		entry, ok := c.entries[key]
		c.mu.Unlock()
		if ok && c.now().Before(entry.expires) {
			return entry.providers, nil
		}
	}

	providers, err := c.cli.ListProviders(ctx, creds, country, currency, destination)
	if err != nil {
		return nil, err
	}
	if c.ttl > 0 {
		c.mu.Lock()
		c.entries[key] = catalogEntry{providers: providers, expires: c.now().Add(c.ttl)}
		c.mu.Unlock()
	}
	return providers, nil
}

// MatchProvider picks the provider named by hint (a code or name, compared
// case-insensitively, also through the network's known aliases). With no
// hint, a corridor served by exactly one provider resolves to it.
func MatchProvider(providers []Provider, hint string) (Provider, bool) {
	hint = normalizeProviderToken(hint)
	if hint == "" {
		if len(providers) == 1 {
			return providers[0], true
		}
		return Provider{}, false
	}
	aliases := providerAliases(hint)
	for _, alias := range aliases {
		for _, p := range providers {
			if normalizeProviderToken(p.Code) == alias || normalizeProviderToken(p.Name) == alias {
				return p, true
			}
		}
	}
	for _, alias := range aliases {
		for _, p := range providers {
			name := normalizeProviderToken(p.Name)
			if alias != "AIRTELTIGO" && strings.Contains(name, "AIRTELTIGO") {
				continue // Airtel or Tigo must not resolve to AirtelTigo.
			}
			if strings.Contains(name, alias) {
				return p, true
			}
		}
	}
	return Provider{}, false
}

// providerAliases lists the names a network goes by, most specific first.
// AirtelTigo is checked before Airtel and Tigo so a Ghana AirtelTigo wallet
// never matches a plain Airtel operator.
func providerAliases(network string) []string {
	switch network {
	case "MPESA", "SAFARICOM", "VODACOM":
		return []string{"MPESA", "SAFARICOM", "VODACOM"}
	case "AIRTELTIGO":
		return []string{"AIRTELTIGO"}
	case "VODAFONE", "TELECEL":
		return []string{"VODAFONE", "TELECEL"}
	}
	return []string{network}
}

// normalizeProviderToken upper-cases and strips separators, so "M-Pesa",
// "AIRTEL_TIGO" and "Airtel Tigo" compare as MPESA and AIRTELTIGO.
func normalizeProviderToken(s string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(strings.TrimSpace(s)) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}
