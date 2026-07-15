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
	"strings"
	"time"

	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/pitabwire/frame/v2/queue"
	"github.com/pitabwire/util"
)

// LifecycleRouteSource looks up integration routes for subscription fan-out.
// Implemented by repository.IntegrationRouteRepository.
type LifecycleRouteSource interface {
	GetByID(ctx context.Context, id string) (*models.IntegrationRoute, error)
	ListForLifecycle(ctx context.Context, partitionID, eventType string) ([]*models.IntegrationRoute, error)
}

// SubscriptionLifecycleEvent is the payload delivered to external entity
// integrators (product apps, entitlement services, CRM) — the billing-side
// analogue of a payment Send/Receive queue message.
type SubscriptionLifecycleEvent struct {
	Event              string         `json:"event"`
	SubscriptionID     string         `json:"subscription_id"`
	ProfileID          string         `json:"profile_id"`
	PlanID             string         `json:"plan_id"`
	CatalogVersionID   string         `json:"catalog_version_id"`
	State              string         `json:"state"`
	Currency           string         `json:"currency"`
	ExternalEntityID   string         `json:"external_entity_id,omitempty"`
	ExternalEntityType string         `json:"external_entity_type,omitempty"`
	InvoiceID          string         `json:"invoice_id,omitempty"`
	PartitionID        string         `json:"partition_id,omitempty"`
	TenantID           string         `json:"tenant_id,omitempty"`
	Data               map[string]any `json:"data,omitempty"`
	OccurredAt         time.Time      `json:"occurred_at"`
}

// SubscriptionLifecycleNotifier fans subscription lifecycle changes out to
// integration routes / default queue — same role as payment.out.route + queue.
//
// Delivery is best-effort: failures are logged and do not roll back billing.
type SubscriptionLifecycleNotifier interface {
	Notify(ctx context.Context, evt SubscriptionLifecycleEvent) error
	// NotifyFromSubscription builds and sends an event for a subscription row.
	NotifyFromSubscription(ctx context.Context, eventType string, sub *models.Subscription, invoiceID string) error
}

// LifecycleNotifierConfig holds optional default topic (when no DB routes match).
type LifecycleNotifierConfig struct {
	// DefaultTopicName is the publisher reference for the global fallback queue.
	DefaultTopicName string
	// DefaultTopicURI is registered once at process start (see main).
	DefaultTopicURI string
}

type subscriptionLifecycleNotifier struct {
	qMan      queue.Manager
	routeRepo LifecycleRouteSource
	cfg       LifecycleNotifierConfig
}

// NewSubscriptionLifecycleNotifier constructs the payment-style lifecycle fan-out.
// qMan and routeRepo may be nil → Notify becomes a no-op (local/dev).
func NewSubscriptionLifecycleNotifier(
	qMan queue.Manager,
	routeRepo LifecycleRouteSource,
	cfg LifecycleNotifierConfig,
) SubscriptionLifecycleNotifier {
	return &subscriptionLifecycleNotifier{
		qMan:      qMan,
		routeRepo: routeRepo,
		cfg:       cfg,
	}
}

// NoopSubscriptionLifecycleNotifier is used when queue integration is disabled.
func NoopSubscriptionLifecycleNotifier() SubscriptionLifecycleNotifier {
	return &subscriptionLifecycleNotifier{}
}

func (n *subscriptionLifecycleNotifier) NotifyFromSubscription(
	ctx context.Context,
	eventType string,
	sub *models.Subscription,
	invoiceID string,
) error {
	if sub == nil {
		return nil
	}
	evt := EventFromSubscription(eventType, sub, invoiceID)
	return n.Notify(ctx, evt)
}

// EventFromSubscription maps a subscription row into a lifecycle event payload.
func EventFromSubscription(
	eventType string,
	sub *models.Subscription,
	invoiceID string,
) SubscriptionLifecycleEvent {
	data := map[string]any{}
	if sub.Data != nil {
		for k, v := range sub.Data {
			data[k] = v
		}
	}
	extID, _ := stringFromData(sub.Data, models.SubDataExternalEntityID)
	extType, _ := stringFromData(sub.Data, models.SubDataExternalEntityType)
	return SubscriptionLifecycleEvent{
		Event:              eventType,
		SubscriptionID:     sub.GetID(),
		ProfileID:          sub.ProfileID,
		PlanID:             sub.PlanID,
		CatalogVersionID:   sub.CatalogVersionID,
		State:              sub.State,
		Currency:           sub.Currency,
		ExternalEntityID:   extID,
		ExternalEntityType: extType,
		InvoiceID:          invoiceID,
		PartitionID:        sub.GetPartitionID(),
		TenantID:           sub.GetTenantID(),
		Data:               data,
		OccurredAt:         time.Now().UTC(),
	}
}

func (n *subscriptionLifecycleNotifier) Notify(
	ctx context.Context,
	evt SubscriptionLifecycleEvent,
) error {
	if n.qMan == nil {
		return nil
	}
	if evt.OccurredAt.IsZero() {
		evt.OccurredAt = time.Now().UTC()
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	refs := n.resolvePublishers(ctx, evt)
	if len(refs) == 0 {
		util.Log(ctx).
			WithField("event", evt.Event).
			WithField("subscription_id", evt.SubscriptionID).
			Debug("no subscription lifecycle routes matched — skip notify")
		return nil
	}

	var firstErr error
	for _, ref := range refs {
		if pubErr := n.qMan.Publish(ctx, ref, payload, map[string]string{
			"event":              evt.Event,
			"subscription_id":    evt.SubscriptionID,
			"partition_id":       evt.PartitionID,
			"external_entity_id": evt.ExternalEntityID,
		}); pubErr != nil {
			util.Log(ctx).
				WithError(pubErr).
				WithField("route_ref", ref).
				WithField("event", evt.Event).
				Warn("could not publish subscription lifecycle event")
			if firstErr == nil {
				firstErr = pubErr
			}
		}
	}
	return firstErr
}

// resolvePublishers returns queue publisher references for this event.
func (n *subscriptionLifecycleNotifier) resolvePublishers(
	ctx context.Context,
	evt SubscriptionLifecycleEvent,
) []string {
	seen := map[string]struct{}{}
	var refs []string
	add := func(ref string) {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			return
		}
		if _, ok := seen[ref]; ok {
			return
		}
		seen[ref] = struct{}{}
		refs = append(refs, ref)
	}

	// 1) Explicit route on the subscription (product pinned a destination).
	if routeID, ok := stringFromDataMap(evt.Data, models.SubDataIntegrationRouteID); ok && n.routeRepo != nil {
		if route, err := n.routeRepo.GetByID(ctx, routeID); err == nil && route != nil {
			if ensureErr := n.ensurePublisher(ctx, route.GetID(), route.URI); ensureErr != nil {
				util.Log(ctx).WithError(ensureErr).WithField("route_id", routeID).
					Warn("could not ensure explicit integration route publisher")
			} else {
				add(route.GetID())
			}
		}
	}

	// 2) Partition-scoped lifecycle routes (payment-style route table).
	if n.routeRepo != nil {
		routes, err := n.routeRepo.ListForLifecycle(ctx, evt.PartitionID, evt.Event)
		if err != nil {
			util.Log(ctx).WithError(err).Warn("could not list integration routes")
		} else {
			for _, route := range routes {
				if route == nil {
					continue
				}
				if ensureErr := n.ensurePublisher(ctx, route.GetID(), route.URI); ensureErr != nil {
					util.Log(ctx).WithError(ensureErr).WithField("route_id", route.GetID()).
						Warn("could not ensure integration route publisher")
					continue
				}
				add(route.GetID())
			}
		}
	}

	// 3) Global default topic (configured at process start).
	if n.cfg.DefaultTopicName != "" {
		add(n.cfg.DefaultTopicName)
	}

	return refs
}

func (n *subscriptionLifecycleNotifier) ensurePublisher(
	ctx context.Context,
	ref, uri string,
) error {
	if n.qMan == nil || ref == "" || uri == "" {
		return nil
	}
	if _, err := n.qMan.GetPublisher(ref); err == nil {
		return nil
	}
	return n.qMan.AddPublisher(ctx, ref, uri)
}

func stringFromData(data map[string]any, key string) (string, bool) {
	return stringFromDataMap(data, key)
}

func stringFromDataMap(data map[string]any, key string) (string, bool) {
	if data == nil {
		return "", false
	}
	v, ok := data[key]
	if !ok || v == nil {
		return "", false
	}
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	return s, true
}
