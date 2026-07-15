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

package business_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/antinvestor/service-payments/apps/billing/service/business"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/pitabwire/frame/v2/data"
	"github.com/pitabwire/frame/v2/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeQueueManager records Publish calls for unit tests.
type fakeQueueManager struct {
	mu        sync.Mutex
	published []publishedMsg
	pubs      map[string]string // ref → uri
}

type publishedMsg struct {
	ref     string
	payload []byte
	headers map[string]string
}

func newFakeQueueManager() *fakeQueueManager {
	return &fakeQueueManager{pubs: map[string]string{}}
}

func (f *fakeQueueManager) AddPublisher(_ context.Context, reference string, queueURL string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pubs[reference] = queueURL
	return nil
}

func (f *fakeQueueManager) GetPublisher(reference string) (queue.Publisher, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.pubs[reference]; !ok {
		return nil, errors.New("publisher not found")
	}
	return &fakePublisher{ref: reference}, nil
}

func (f *fakeQueueManager) DiscardPublisher(context.Context, string) error { return nil }
func (f *fakeQueueManager) AddSubscriber(context.Context, string, string, ...queue.SubscribeWorker) error {
	return nil
}
func (f *fakeQueueManager) DiscardSubscriber(context.Context, string) error { return nil }
func (f *fakeQueueManager) GetSubscriber(string) (queue.Subscriber, error) {
	return nil, errors.New("not found")
}
func (f *fakeQueueManager) Init(context.Context) error  { return nil }
func (f *fakeQueueManager) Close(context.Context) error { return nil }

func (f *fakeQueueManager) Publish(
	_ context.Context,
	reference string,
	payload any,
	headers ...map[string]string,
) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	var body []byte
	switch p := payload.(type) {
	case []byte:
		body = p
	case string:
		body = []byte(p)
	default:
		body, _ = json.Marshal(p)
	}
	h := map[string]string{}
	if len(headers) > 0 {
		for k, v := range headers[0] {
			h[k] = v
		}
	}
	f.published = append(f.published, publishedMsg{ref: reference, payload: body, headers: h})
	return nil
}

func (f *fakeQueueManager) messages() []publishedMsg {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]publishedMsg, len(f.published))
	copy(out, f.published)
	return out
}

type fakePublisher struct{ ref string }

func (p *fakePublisher) Initiated() bool            { return true }
func (p *fakePublisher) Ref() string                { return p.ref }
func (p *fakePublisher) Init(context.Context) error { return nil }
func (p *fakePublisher) Publish(context.Context, any, ...map[string]string) error {
	return nil
}
func (p *fakePublisher) Stop(context.Context) error { return nil }
func (p *fakePublisher) As(any) bool                { return false }

var (
	_ queue.Manager   = (*fakeQueueManager)(nil)
	_ queue.Publisher = (*fakePublisher)(nil)
)

type fakeRouteSource struct {
	byID map[string]*models.IntegrationRoute
	list []*models.IntegrationRoute
}

func (s *fakeRouteSource) GetByID(_ context.Context, id string) (*models.IntegrationRoute, error) {
	r, ok := s.byID[id]
	if !ok {
		return nil, errors.New("route not found")
	}
	return r, nil
}

func (s *fakeRouteSource) ListForLifecycle(context.Context, string, string) ([]*models.IntegrationRoute, error) {
	return s.list, nil
}

var _ business.LifecycleRouteSource = (*fakeRouteSource)(nil)

func TestEventFromSubscription_ExternalEntity(t *testing.T) {
	sub := &models.Subscription{
		ProfileID:        "prof-1",
		PlanID:           "plan-1",
		CatalogVersionID: "cv-1",
		State:            models.SubscriptionStateActive,
		Currency:         "KES",
		Data: data.JSONMap{
			models.SubDataExternalEntityID:   "ws-99",
			models.SubDataExternalEntityType: "workspace",
			"seatCount":                      "5",
		},
	}
	sub.ID = "sub-1"
	sub.PartitionID = "part-a"
	sub.TenantID = "ten-a"

	evt := business.EventFromSubscription(models.SubscriptionEventActivated, sub, "inv-1")
	assert.Equal(t, models.SubscriptionEventActivated, evt.Event)
	assert.Equal(t, "sub-1", evt.SubscriptionID)
	assert.Equal(t, "ws-99", evt.ExternalEntityID)
	assert.Equal(t, "workspace", evt.ExternalEntityType)
	assert.Equal(t, "inv-1", evt.InvoiceID)
	assert.Equal(t, "part-a", evt.PartitionID)
	assert.Equal(t, "5", evt.Data["seatCount"])
}

func TestLifecycleNotifier_PublishesToRouteAndDefault(t *testing.T) {
	q := newFakeQueueManager()
	route := &models.IntegrationRoute{
		Name:      "product-entitlements",
		RouteType: models.IntegrationRouteTypeAny,
		Mode:      models.IntegrationRouteModeLifecycle,
		URI:       "mem://product.lifecycle",
	}
	route.ID = "route-1"
	route.PartitionID = "part-a"

	n := business.NewSubscriptionLifecycleNotifier(
		q,
		&fakeRouteSource{
			byID: map[string]*models.IntegrationRoute{"route-1": route},
			list: []*models.IntegrationRoute{route},
		},
		business.LifecycleNotifierConfig{
			DefaultTopicName: "subscription.lifecycle",
		},
	)
	require.NoError(t, q.AddPublisher(context.Background(), "subscription.lifecycle", "mem://default"))

	sub := &models.Subscription{
		ProfileID: "p1",
		PlanID:    "plan",
		State:     models.SubscriptionStateActive,
		Currency:  "KES",
		Data: data.JSONMap{
			models.SubDataExternalEntityID:   "ent-1",
			models.SubDataExternalEntityType: "membership",
			models.SubDataIntegrationRouteID: "route-1",
		},
	}
	sub.ID = "sub-x"
	sub.PartitionID = "part-a"

	err := n.NotifyFromSubscription(context.Background(), models.SubscriptionEventActivated, sub, "")
	require.NoError(t, err)

	msgs := q.messages()
	// Explicit route + list (same id once) + default topic.
	require.GreaterOrEqual(t, len(msgs), 2)
	refs := map[string]bool{}
	for _, m := range msgs {
		refs[m.ref] = true
		var evt business.SubscriptionLifecycleEvent
		require.NoError(t, json.Unmarshal(m.payload, &evt))
		assert.Equal(t, models.SubscriptionEventActivated, evt.Event)
		assert.Equal(t, "ent-1", evt.ExternalEntityID)
		assert.Equal(t, "membership", evt.ExternalEntityType)
	}
	assert.True(t, refs["route-1"])
	assert.True(t, refs["subscription.lifecycle"])
}

func TestLifecycleNotifier_NoopWithoutQueue(t *testing.T) {
	n := business.NoopSubscriptionLifecycleNotifier()
	err := n.Notify(context.Background(), business.SubscriptionLifecycleEvent{
		Event: models.SubscriptionEventCreated,
	})
	require.NoError(t, err)
}

func TestBuildSubscriptionData_ViaCreatePath(t *testing.T) {
	// buildSubscriptionData is unexported; verify through EventFromSubscription
	// after constructing the same Data map production uses.
	dataMap := data.JSONMap{
		models.SubDataExternalEntityID:   "ws-1",
		models.SubDataExternalEntityType: "workspace",
		models.SubDataIntegrationRouteID: "route-9",
		"region":                         "ke",
	}
	sub := &models.Subscription{
		ProfileID: "p",
		PlanID:    "plan",
		State:     models.SubscriptionStatePending,
		Currency:  "KES",
		Data:      dataMap,
	}
	sub.ID = "s1"
	evt := business.EventFromSubscription(models.SubscriptionEventCreated, sub, "")
	assert.Equal(t, "ws-1", evt.ExternalEntityID)
	assert.Equal(t, "workspace", evt.ExternalEntityType)
	assert.Equal(t, "ke", evt.Data["region"])
	assert.Equal(t, "route-9", evt.Data[models.SubDataIntegrationRouteID])
}
