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
	"math"
	"strconv"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/pkg/events"
	frameEvents "github.com/pitabwire/frame/v2/events"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/types/known/structpb"
)

type statusEmitter struct {
	eventsMan frameEvents.Manager
}

func (e *statusEmitter) emitStatus(
	ctx context.Context,
	id, externalID string,
	status commonv1.STATUS,
	extras map[string]any,
) {
	if extras == nil {
		extras = map[string]any{}
	}
	extra, _ := structpb.NewStruct(extras)
	state := commonv1.STATE_ACTIVE
	if status == commonv1.STATUS_FAILED {
		state = commonv1.STATE_INACTIVE
	}
	err := e.eventsMan.Emit(ctx, events.PaymentStatusUpdateEvent, &commonv1.StatusUpdateRequest{
		Id:         id,
		State:      state,
		Status:     status,
		ExternalId: externalID,
		Extras:     extra,
	})
	if err != nil {
		util.Log(ctx).WithError(err).Warn("could not emit status update")
	}
}

// formatMoneyAmount converts google.type.Money to a decimal string for Flutterwave.
func formatMoneyAmount(amount interface {
	GetUnits() int64
	GetNanos() int32
	GetCurrencyCode() string
}) (string, string) {
	if amount == nil {
		return "0", ""
	}
	currency := strings.ToUpper(amount.GetCurrencyCode())
	units := amount.GetUnits()
	nanos := amount.GetNanos()
	// Preserve minor units when present (e.g. 10.50).
	if nanos == 0 {
		return strconv.FormatInt(units, 10), currency
	}
	total := float64(units) + float64(nanos)/1e9 //nolint:mnd // nanos conversion
	// Flutterwave expects plain numbers; avoid scientific notation.
	return strconv.FormatFloat(total, 'f', -1, 64), currency
}

// moneyToWholeUnits rounds Money to whole currency units for MoMo transfers that require integers.
func moneyToWholeUnits(amount interface {
	GetUnits() int64
	GetNanos() int32
}) int64 {
	if amount == nil {
		return 0
	}
	total := float64(amount.GetUnits()) + float64(amount.GetNanos())/1e9 //nolint:mnd // nanos
	return int64(math.Round(total))
}

func headerOrDefault(headers map[string]string, key, fallback string) string {
	if v, ok := headers[key]; ok && v != "" {
		return v
	}
	return fallback
}

func extraString(extra *structpb.Struct, key string) string {
	if extra == nil {
		return ""
	}
	f, ok := extra.GetFields()[key]
	if !ok || f == nil {
		return ""
	}
	return f.GetStringValue()
}
