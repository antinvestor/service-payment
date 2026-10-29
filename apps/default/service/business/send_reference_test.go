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
	"strings"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"connectrpc.com/connect/v2"
	"github.com/pitabwire/frame/v3/security"
	utilmoney "github.com/pitabwire/util/moneyx"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"

	"github.com/pitabwire/frame/v3/data"

	"github.com/antinvestor/service-payments/apps/default/service/metrics"
	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
)

// paymentsByReference holds outbound payments keyed by reference.
type paymentsByReference struct {
	repository.PaymentRepository

	byReference map[string]*models.Payment
}

func (r *paymentsByReference) GetOutboundByReference(
	_ context.Context,
	_, _, ref string,
) (*models.Payment, error) {
	if p, ok := r.byReference[ref]; ok {
		return p, nil
	}
	return nil, gorm.ErrRecordNotFound
}

// statusesByEntity answers each payment's latest status.
type statusesByEntity struct {
	repository.StatusRepository

	byEntity map[string]*models.Status
}

func (r *statusesByEntity) Create(_ context.Context, status *models.Status) error {
	for _, saved := range r.byEntity {
		if saved.ID == status.ID {
			return gorm.ErrDuplicatedKey
		}
	}
	r.byEntity[status.EntityID] = status
	return nil
}

func (r *statusesByEntity) GetByID(_ context.Context, id string) (*models.Status, error) {
	for _, saved := range r.byEntity {
		if saved.ID == id {
			return saved, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *statusesByEntity) GetByEntity(_ context.Context, id, _ string) (*models.Status, error) {
	if s, ok := r.byEntity[id]; ok {
		return s, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func TestSendWithAKnownReferenceAnswersTheFirstPayment(t *testing.T) {
	amount := utilmoney.FromMoney(&commonv1.Money{CurrencyCode: "KES", Units: 250})
	prior := &models.Payment{
		ReferenceID:        "refund-1",
		OutBound:           true,
		RecipientProfileID: "buyer-1",
		Currency:           "KES",
		Amount:             &amount,
	}
	prior.ID = "payment-1"
	status := &models.Status{
		EntityID:   prior.ID,
		EntityType: "payment",
		State:      int32(commonv1.STATE_ACTIVE.Number()),
		Status:     int32(commonv1.STATUS_SUCCESSFUL.Number()),
	}
	pb := &paymentBusiness{
		paymentRepo: &paymentsByReference{byReference: map[string]*models.Payment{"refund-1": prior}},
		statusRepo:  &statusesByEntity{byEntity: map[string]*models.Status{prior.ID: status}},
		obs:         metrics.NewMetrics(),
	}

	resend := &paymentv1.Payment{
		ReferenceId: "refund-1",
		Recipient:   &commonv1.ContactLink{ProfileId: "buyer-1"},
		Amount:      &commonv1.Money{CurrencyCode: "KES", Units: 250},
	}
	// No profile client is wired: a resend must not resolve the payee again.
	resp, err := pb.Send(context.Background(), resend)
	require.NoError(t, err)
	require.Equal(t, commonv1.STATUS_SUCCESSFUL, resp.GetStatus())

	// The same reference for someone else, or another amount, is refused.
	other := proto.Clone(resend).(*paymentv1.Payment)
	other.Recipient = &commonv1.ContactLink{ProfileId: "someone-else"}
	_, err = pb.Send(context.Background(), other)
	require.Equal(t, connect.CodeAlreadyExists, connect.CodeOf(err))

	more := proto.Clone(resend).(*paymentv1.Payment)
	more.Amount = &commonv1.Money{CurrencyCode: "KES", Units: 251}
	_, err = pb.Send(context.Background(), more)
	require.Equal(t, connect.CodeAlreadyExists, connect.CodeOf(err))
}

func TestReferencedIDsDifferByPartition(t *testing.T) {
	inPartition := func(partition string) context.Context {
		claims := &security.AuthenticationClaims{TenantID: "tenant-1", PartitionID: partition}
		return claims.ClaimsToContext(context.Background())
	}
	a := referencedID(inPartition("partition-a"), "refund-1")
	require.Equal(t, a, referencedID(inPartition("partition-a"), "refund-1"))
	require.NotEqual(t, a, referencedID(inPartition("partition-b"), "refund-1"))
	require.Empty(t, referencedID(inPartition("partition-a"), ""))
}

func TestSendRefusesAReferenceTooLongToStore(t *testing.T) {
	pb := &paymentBusiness{obs: metrics.NewMetrics()}
	_, err := pb.Send(context.Background(), &paymentv1.Payment{ReferenceId: strings.Repeat("r", 151)})
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestClaimReference(t *testing.T) {
	send := &paymentv1.Payment{
		ReferenceId: "refund-1",
		Recipient:   &commonv1.ContactLink{ProfileId: "buyer-1"},
		Amount:      &commonv1.Money{CurrencyCode: "KES", Units: 250},
	}
	claim := func(id string) *models.Status {
		s := &models.Status{EntityID: "payment-1", EntityType: "payment", Extra: data.JSONMap{}}
		s.ID = id
		return s
	}
	claimedBy := func(m *paymentv1.Payment) *models.Status {
		s := claim("ref-queued")
		s.Extra[extraSendFingerprint] = sendFingerprint(m)
		return s
	}
	elsewhere := proto.Clone(send).(*paymentv1.Payment)
	elsewhere.Recipient = &commonv1.ContactLink{ProfileId: "someone-else"}

	t.Run("the first send takes the claim", func(t *testing.T) {
		pb := &paymentBusiness{
			paymentRepo: &paymentsByReference{byReference: map[string]*models.Payment{}},
			statusRepo:  &statusesByEntity{byEntity: map[string]*models.Status{}},
		}
		resp, err := pb.claimReference(context.Background(), send, claim("ref-queued"))
		require.NoError(t, err)
		require.Nil(t, resp, "the claimant goes on sending")
	})

	t.Run("a racing send for someone else is refused before the payment is saved", func(t *testing.T) {
		pb := &paymentBusiness{
			paymentRepo: &paymentsByReference{byReference: map[string]*models.Payment{}},
			statusRepo:  &statusesByEntity{byEntity: map[string]*models.Status{"payment-1": claimedBy(elsewhere)}},
		}
		_, err := pb.claimReference(context.Background(), send, claim("ref-queued"))
		require.Equal(t, connect.CodeAlreadyExists, connect.CodeOf(err))
	})

	t.Run("the same send racing, before the payment is saved, saves the same payment", func(t *testing.T) {
		pb := &paymentBusiness{
			paymentRepo: &paymentsByReference{byReference: map[string]*models.Payment{}},
			statusRepo:  &statusesByEntity{byEntity: map[string]*models.Status{"payment-1": claimedBy(send)}},
		}
		resp, err := pb.claimReference(context.Background(), send, claim("ref-queued"))
		require.NoError(t, err)
		require.Nil(t, resp)
	})
}
