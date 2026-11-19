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

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/apps/integrations/pawapay/service/client"
	"github.com/antinvestor/service-payments/pkg/events"
	"github.com/antinvestor/service-payments/pkg/integrationobs"
)

const providerName = "pawapay"

// recordStatus maps a pawaPay deposit/payout status to the platform status.
func recordStatus(status string) commonv1.STATUS {
	switch status {
	case client.PaymentStatusCompleted:
		return commonv1.STATUS_SUCCESSFUL
	case client.PaymentStatusFailed:
		return commonv1.STATUS_FAILED
	default:
		return commonv1.STATUS_IN_PROCESS
	}
}

// reporter emits the status of one pawaPay record under our entity id.
type reporter struct {
	emitter    *events.StatusEmitter
	metrics    *integrationobs.Metrics
	entityType string // events.EntityTypePrompt or events.EntityTypePayment
	kind       string // deposit or payout: names the provider id extra
}

// fail reports a FAILED status with the given metrics reason.
func (r *reporter) fail(ctx context.Context, id, externalID, reason string, extras map[string]any) error {
	r.metrics.QueueFailed(ctx, r.entityType, reason)
	if extras == nil {
		extras = map[string]any{}
	}
	extras["provider"] = providerName
	if externalID != "" {
		extras[r.kind+"_id"] = externalID
	}
	return r.emitter.Emit(ctx, r.entityType, id, externalID, commonv1.STATUS_FAILED, extras)
}

// record reports what pawaPay holds for an initiated deposit or payout.
func (r *reporter) record(
	ctx context.Context,
	id, providerID, status, provider, created string,
	failure *client.FailureReason,
) error {
	extras := map[string]any{
		r.kind + "_id": providerID,
		"provider":     providerName,
		"status":       status,
	}
	if provider != "" {
		extras["network"] = provider
	}
	if created != "" {
		extras["created"] = created
	}
	mapped := recordStatus(status)
	if mapped == commonv1.STATUS_FAILED {
		extras["failure_code"] = failureCode(failure)
		extras["failure_message"] = failureMessage(failure)
		return r.fail(ctx, id, providerID, "rejected", extras)
	}
	r.metrics.QueueProcessed(ctx, r.entityType)
	return r.emitter.Emit(ctx, r.entityType, id, providerID, mapped, extras)
}
