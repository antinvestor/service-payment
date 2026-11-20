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
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/pitabwire/frame/v3/queue"
	"github.com/pitabwire/util"
)

// ErrNoProviderCommandRoute is returned when no command queue is configured
// for a provider.
var ErrNoProviderCommandRoute = errors.New("no command queue configured for provider")

// ProviderCommandPublisher sends commands (cancel, …) to the integration
// worker that talks to a provider, over that provider's command queue.
type ProviderCommandPublisher interface {
	Publish(ctx context.Context, cmd collection.ProviderCommand) error
	// Configured reports whether a command queue exists for the provider.
	Configured(provider string) bool
}

// ParseProviderCommandURIs parses BILLING_PROVIDER_COMMAND_URIS:
// {"polar":"nats://…/polar.subscriptions.dequeue"}.
func ParseProviderCommandURIs(raw string) (map[string]string, error) {
	out := map[string]string{}
	if strings.TrimSpace(raw) == "" {
		return out, nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, fmt.Errorf("parse provider command uris: %w", err)
	}
	for k, v := range m {
		k = strings.ToLower(strings.TrimSpace(k))
		v = strings.TrimSpace(v)
		if k == "" || v == "" {
			continue
		}
		out[k] = v
	}
	return out, nil
}

type providerCommandPublisher struct {
	qMan      queue.Manager
	providers map[string]bool
}

// NewProviderCommandPublisher publishes on the publishers main registered
// (collection.CommandPublisherRef) for each configured provider.
func NewProviderCommandPublisher(qMan queue.Manager, uris map[string]string) ProviderCommandPublisher {
	providers := make(map[string]bool, len(uris))
	for p := range uris {
		providers[p] = true
	}
	return &providerCommandPublisher{qMan: qMan, providers: providers}
}

// NoopProviderCommandPublisher never publishes (local / tests).
func NoopProviderCommandPublisher() ProviderCommandPublisher {
	return &providerCommandPublisher{}
}

func (p *providerCommandPublisher) Configured(provider string) bool {
	return p != nil && p.qMan != nil && p.providers[strings.ToLower(provider)]
}

func (p *providerCommandPublisher) Publish(ctx context.Context, cmd collection.ProviderCommand) error {
	cmd.Provider = strings.ToLower(strings.TrimSpace(cmd.Provider))
	if !p.Configured(cmd.Provider) {
		return fmt.Errorf("%w: %s", ErrNoProviderCommandRoute, cmd.Provider)
	}
	if cmd.IssuedAt.IsZero() {
		cmd.IssuedAt = time.Now().UTC()
	}
	payload, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("encode provider command: %w", err)
	}
	headers := map[string]string{
		"provider":        cmd.Provider,
		"command":         cmd.Command,
		"subscription_id": cmd.SubscriptionID,
	}
	if err = p.qMan.Publish(ctx, collection.CommandPublisherRef(cmd.Provider), payload, headers); err != nil {
		return fmt.Errorf("publish provider command: %w", err)
	}
	util.Log(ctx).WithField("provider", cmd.Provider).WithField("command", cmd.Command).
		WithField("subscription_id", cmd.SubscriptionID).Info("provider command published")
	return nil
}

// cancelCommandFor builds the cancel command for a provider-managed subscription.
func cancelCommandFor(sub *models.Subscription, atPeriodEnd bool) (collection.ProviderCommand, bool) {
	provider := SubscriptionProvider(sub)
	externalID := ExternalSubscriptionID(sub)
	if provider == "" || externalID == "" {
		return collection.ProviderCommand{}, false
	}
	return collection.ProviderCommand{
		Provider:               provider,
		Command:                collection.CommandCancelSubscription,
		SubscriptionID:         sub.GetID(),
		ExternalSubscriptionID: externalID,
		AtPeriodEnd:            atPeriodEnd,
		TenantID:               sub.GetTenantID(),
		PartitionID:            sub.GetPartitionID(),
	}, true
}
