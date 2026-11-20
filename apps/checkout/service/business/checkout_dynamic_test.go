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
	"testing"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/apps/checkout/service/business"
	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// Pay enforces the partition allowlist the page applied.
func TestPay_PartitionAllowlistEnforced(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	payCli := &fakePaymentClient{}
	b := newBusiness(defaultConfig(), defaultRegistry(), sessionRepo, newFakeLinkRepo(), payCli, &fakeProfileClient{}).
		WithPartitionAllowlists(business.PartitionAllowlists{"part-a": {"mtn_momo"}})

	sess := &models.CheckoutSession{
		Ref:          "sess-part",
		Status:       models.SessionStatusPending,
		ExpiresAt:    fixedNow().Add(20 * time.Minute),
		Currency:     "KES",
		Amount:       "100.00",
		AmountOption: models.AmountOptionFixed,
	}
	sess.PartitionID = "part-a"
	sessionRepo.sessions[sess.Ref] = sess

	_, err := b.Pay(context.Background(), sess.Ref, business.PayInput{MethodKey: "mpesa", PhoneNumber: "254700000001"})
	require.ErrorIs(t, err, business.ErrUnknownMethod)
	assert.Nil(t, payCli.lastPrompt, "no prompt may be sent for a partition-excluded method")
	assert.Equal(t, []string{"mtn_momo"}, b.PartitionAllowlist("part-a"))
	assert.Nil(t, b.PartitionAllowlist("other"))
}

// The prompt carries provider-portable extras: product reference, cancel
// URL, stable payer id and the instrument class.
func TestPay_PromptCarriesPortableExtras(t *testing.T) {
	reg, err := business.BuildMethodRegistry("fincra,polar", "")
	require.NoError(t, err)
	sessionRepo := newFakeSessionRepo()
	payCli := &fakePaymentClient{}
	b := newBusiness(defaultConfig(), reg, sessionRepo, newFakeLinkRepo(), payCli, &fakeProfileClient{})

	sess := &models.CheckoutSession{
		Ref:            "sess-polar",
		Status:         models.SessionStatusPending,
		ExpiresAt:      fixedNow().Add(20 * time.Minute),
		Currency:       "USD",
		Amount:         "10.00",
		AmountOption:   models.AmountOptionFixed,
		OrderRef:       "inv-1",
		PayerProfileID: "prof-1",
		Metadata:       map[string]any{"polarProductId": "prod_123", "subscriptionId": "sub-9", "source": "billing"},
		Prefill:        map[string]any{"email": "payer@example.com"},
	}
	sessionRepo.sessions[sess.Ref] = sess

	_, err = b.Pay(context.Background(), sess.Ref, business.PayInput{MethodKey: "polar", ContactID: "", Email: "payer@example.com"})
	require.NoError(t, err)
	require.NotNil(t, payCli.lastPrompt)
	fields := payCli.lastPrompt.GetExtra().GetFields()
	assert.Equal(t, "polar", payCli.lastPrompt.GetRoute())
	assert.Equal(t, "prod_123", fields[collection.ExtraProductID].GetStringValue())
	assert.Equal(t, "sub-9", fields[collection.ExtraSubscriptionID].GetStringValue())
	assert.Equal(t, "billing", fields[collection.ExtraCollectionSource].GetStringValue())
	assert.Equal(t, "prof-1", fields[collection.ExtraCustomerExternalID].GetStringValue())
	assert.Equal(t, collection.PaymentMethodTypeCard, fields[collection.ExtraPaymentMethodType].GetStringValue())
	assert.Equal(t, "http://localhost:8080/c/sess-polar", fields[collection.ExtraSuccessURL].GetStringValue())
	assert.Equal(t, "http://localhost:8080/c/sess-polar?cancelled=1", fields[collection.ExtraCancelURL].GetStringValue())
	assert.Equal(t, "payer@example.com", fields[collection.ExtraCustomerEmail].GetStringValue())
}

// A session restricted to a product rail must carry the product reference.
func TestCreateSession_ProductRailNeedsProductID(t *testing.T) {
	reg, err := business.BuildMethodRegistry("fincra,polar", "")
	require.NoError(t, err)
	b := newBusiness(defaultConfig(), reg, newFakeSessionRepo(), newFakeLinkRepo(), &fakePaymentClient{}, &fakeProfileClient{})

	in := business.CreateSessionInput{Name: "Plan", Amount: "10.00", Currency: "USD", Methods: []string{"polar"}}
	_, err = b.CreateSession(context.Background(), in)
	require.ErrorIs(t, err, business.ErrUnknownMethod)

	in.Metadata = map[string]string{"product_id": "prod_1"}
	in.CancelURL = "javascript:alert(1)"
	_, err = b.CreateSession(context.Background(), in)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancel_url")

	in.CancelURL = "https://app.example.com/cancel"
	sess, err := b.CreateSession(context.Background(), in)
	require.NoError(t, err)
	assert.Equal(t, "https://app.example.com/cancel", sess.CancelURL)
	assert.Equal(t, "prod_1", business.SessionProductID(sess))
}

func TestRefreshStatus_CapturesClientSecretAndExternalSubscription(t *testing.T) {
	sessionRepo := newFakeSessionRepo()
	extras, err := structpb.NewStruct(map[string]any{
		collection.ExtraClientSecret:           "pi_secret",
		collection.ExtraExternalSubscriptionID: "sub_ext",
	})
	require.NoError(t, err)
	payCli := &fakePaymentClient{statusByID: map[string]*commonv1.StatusResponse{
		"prompt-1": {Id: "prompt-1", Status: commonv1.STATUS_IN_PROCESS, Extras: extras},
	}}
	b := newBusiness(defaultConfig(), defaultRegistry(), sessionRepo, newFakeLinkRepo(), payCli, &fakeProfileClient{})
	sess := processingSession("sess-cs", "prompt-1", fixedNow(), "prompt-1")
	sessionRepo.sessions[sess.Ref] = sess

	refreshed, err := b.RefreshStatus(context.Background(), sess)
	require.NoError(t, err)
	assert.Equal(t, models.SessionStatusProcessing, refreshed.Status)
	assert.Equal(t, "pi_secret", refreshed.Metadata["_client_secret"])
	assert.Equal(t, "sub_ext", refreshed.Metadata["_external_subscription_id"])
}
