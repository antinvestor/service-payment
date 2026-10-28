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
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/antinvestor/service-payments/apps/default/service/metrics"
	"github.com/antinvestor/service-payments/apps/default/service/models"
	"github.com/antinvestor/service-payments/apps/default/service/repository"
)

// paymentsByReference holds outbound payments keyed by reference.
type paymentsByReference struct {
	repository.PaymentRepository

	byReference map[string]*models.Payment
}

func (r *paymentsByReference) GetOutboundByReference(_ context.Context, ref string) (*models.Payment, error) {
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

func (r *statusesByEntity) GetByEntity(_ context.Context, id, _ string) (*models.Status, error) {
	if s, ok := r.byEntity[id]; ok {
		return s, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func TestSendWithAKnownReferenceAnswersTheFirstPayment(t *testing.T) {
	prior := &models.Payment{ReferenceID: "refund-1", OutBound: true}
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

	// No profile client is wired: a resend must not resolve the payee again.
	resp, err := pb.Send(context.Background(), &paymentv1.Payment{ReferenceId: "refund-1"})
	require.NoError(t, err)
	require.Equal(t, commonv1.STATUS_SUCCESSFUL, resp.GetStatus())
}

func TestSendRefusesAReferenceTooLongToStore(t *testing.T) {
	pb := &paymentBusiness{obs: metrics.NewMetrics()}
	_, err := pb.Send(context.Background(), &paymentv1.Payment{ReferenceId: strings.Repeat("r", 151)})
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}
