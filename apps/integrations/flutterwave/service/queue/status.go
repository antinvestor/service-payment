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
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"github.com/antinvestor/service-payments/pkg/events"
	frameEvents "github.com/pitabwire/frame/v3/events"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/types/known/structpb"
)

type statusEmitter struct {
	eventsMan frameEvents.Manager
}

// emitStatus reports an entity's status to the payment service. The error
// is returned so a worker can redeliver rather than lose an outcome.
func (e *statusEmitter) emitStatus(
	ctx context.Context,
	id, externalID string,
	status commonv1.STATUS,
	extras map[string]any,
) error {
	if extras == nil {
		extras = map[string]any{}
	}
	extra, err := structpb.NewStruct(extras)
	if err != nil {
		return fmt.Errorf("status extras: %w", err)
	}
	state := commonv1.STATE_ACTIVE
	if status == commonv1.STATUS_FAILED {
		state = commonv1.STATE_INACTIVE
	}
	err = e.eventsMan.Emit(ctx, events.PaymentStatusUpdateEvent, &commonv1.StatusUpdateRequest{
		Id:         id,
		State:      state,
		Status:     status,
		ExternalId: externalID,
		Extras:     extra,
	})
	if err != nil {
		util.Log(ctx).WithError(err).WithField("entity_id", id).Warn("could not emit status update")
		return err
	}
	return nil
}

// wholeUnitCurrencies have no minor unit in practice: an amount with a
// fraction cannot be charged or paid exactly.
var wholeUnitCurrencies = map[string]bool{
	"UGX": true, "RWF": true, "XOF": true, "XAF": true, "BIF": true, "GNF": true,
	"KMF": true, "DJF": true, "JPY": true, "KRW": true, "VND": true, "CLP": true, "PYG": true,
}

// errInvalidAmount is an amount Flutterwave cannot be asked for exactly.
var errInvalidAmount = errors.New(
	"amount must be positive with at most two decimal places (whole units for UGX, RWF, XOF, XAF and similar)")

const nanosPerCent = 10_000_000

// providerAmount is the exact amount and currency to send. An amount that
// cannot be sent exactly is refused, never rounded.
func providerAmount(amount interface {
	GetUnits() int64
	GetNanos() int32
	GetCurrencyCode() string
}) (float64, string, error) {
	if amount == nil {
		return 0, "", errInvalidAmount
	}
	currency := strings.ToUpper(strings.TrimSpace(amount.GetCurrencyCode()))
	units, nanos := amount.GetUnits(), amount.GetNanos()
	if currency == "" {
		return 0, "", errors.New("amount has no currency")
	}
	if units < 0 || nanos < 0 || (units == 0 && nanos == 0) || nanos%nanosPerCent != 0 {
		return 0, currency, errInvalidAmount
	}
	if nanos != 0 && wholeUnitCurrencies[currency] {
		return 0, currency, errInvalidAmount
	}
	cents := units*100 + int64(nanos/nanosPerCent)                                                          //nolint:mnd // cents
	value, _ := strconv.ParseFloat(strconv.FormatInt(cents/100, 10)+"."+fmt.Sprintf("%02d", cents%100), 64) //nolint:mnd // cents
	return value, currency, nil
}

// sameAmount compares provider amounts to the cent.
func sameAmount(a, b float64) bool {
	return math.Round(a*100) == math.Round(b*100) //nolint:mnd // cents
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
