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

package events

import (
	"context"
	"fmt"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/types/known/structpb"
)

// EntityTypePrompt and EntityTypePayment are the status entity kinds the
// payment service keys status rows by.
const (
	EntityTypePrompt  = "prompt"
	EntityTypePayment = "payment"
	// ExtraEntityType names the extras key carrying the entity kind.
	ExtraEntityType = "entity_type"
	// ExtraError carries a provider's failure message on FAILED statuses.
	ExtraError = "error"
)

// StateFor is the platform state that goes with a status: FAILED records are
// inactive, everything else stays active.
func StateFor(status commonv1.STATUS) commonv1.STATE {
	if status == commonv1.STATUS_FAILED {
		return commonv1.STATE_INACTIVE
	}
	return commonv1.STATE_ACTIVE
}

// StatusEmitter publishes payment status updates from an integration worker.
type StatusEmitter struct {
	eventsMan frameEvents.Manager
}

// NewStatusEmitter wraps a frame events manager.
func NewStatusEmitter(eventsMan frameEvents.Manager) *StatusEmitter {
	return &StatusEmitter{eventsMan: eventsMan}
}

// Emit publishes a StatusUpdateRequest for id. The error is returned so the
// worker can redeliver the message rather than lose the status; workers that
// look their reference up first make that redelivery safe.
func (e *StatusEmitter) Emit(
	ctx context.Context,
	entityType, id, externalID string,
	status commonv1.STATUS,
	extras map[string]any,
) error {
	if e == nil || e.eventsMan == nil {
		return fmt.Errorf("status emitter not configured")
	}
	if extras == nil {
		extras = map[string]any{}
	}
	extras[ExtraEntityType] = entityType
	extra, err := structpb.NewStruct(extras)
	if err != nil {
		util.Log(ctx).WithError(err).Warn("could not encode status extras; sending entity type only")
		extra, _ = structpb.NewStruct(map[string]any{ExtraEntityType: entityType})
	}
	err = e.eventsMan.Emit(ctx, PaymentStatusUpdateEvent, &commonv1.StatusUpdateRequest{
		Id:         id,
		State:      StateFor(status),
		Status:     status,
		ExternalId: externalID,
		Extras:     extra,
	})
	if err != nil {
		util.Log(ctx).WithError(err).WithField("id", id).Warn("could not emit status update")
		return fmt.Errorf("emit status for %s: %w", id, err)
	}
	return nil
}

// Fail emits a FAILED status with the provider's reason.
func (e *StatusEmitter) Fail(ctx context.Context, entityType, id, reason string, extras map[string]any) error {
	if extras == nil {
		extras = map[string]any{}
	}
	extras[ExtraError] = reason
	return e.Emit(ctx, entityType, id, "", commonv1.STATUS_FAILED, extras)
}
