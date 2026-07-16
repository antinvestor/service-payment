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
	"fmt"
	"strings"
	"time"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	paymentv1 "buf.build/gen/go/antinvestor/payment/protocolbuffers/go/v1"
	"buf.build/gen/go/antinvestor/profile/connectrpc/go/profile/v1/profilev1connect"
	profilev1 "buf.build/gen/go/antinvestor/profile/protocolbuffers/go/profile/v1"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/billing/service/models"
	"github.com/antinvestor/service-payments/pkg/collection"
	"github.com/pitabwire/util"
	"google.golang.org/protobuf/types/known/structpb"
)

// SavedInstrument is a Flutterwave v4 (or portable) card-on-file reference.
type SavedInstrument struct {
	PaymentMethodID string
	CustomerID      string
	Provider        string // e.g. flutterwave
}

// InstrumentSource resolves COF credentials for silent renewals.
type InstrumentSource interface {
	ResolveInstrument(ctx context.Context, sub *models.Subscription) (*SavedInstrument, error)
}

// PaymentCollector charges an issued invoice off-session (no hosted page).
// Implementation uses payment InitiatePrompt with payment_method_id — Flutterwave v4 only.
type PaymentCollector interface {
	// CollectCOF starts a token charge for the invoice. Returns prompt id for polling.
	CollectCOF(ctx context.Context, invoice *models.Invoice, inst *SavedInstrument, route string) (promptID string, err error)
	// PollPromptStatus returns SUCCESSFUL/FAILED/IN_PROCESS for a collection prompt.
	PollPromptStatus(ctx context.Context, promptID string) (commonv1.STATUS, error)
}

type paymentCollector struct {
	pay paymentv1connect.PaymentServiceClient
}

// NewPaymentCollector builds a COF collector. pay may be nil → methods error.
func NewPaymentCollector(pay paymentv1connect.PaymentServiceClient) PaymentCollector {
	return &paymentCollector{pay: pay}
}

func (c *paymentCollector) CollectCOF(
	ctx context.Context,
	invoice *models.Invoice,
	inst *SavedInstrument,
	route string,
) (string, error) {
	if c == nil || c.pay == nil {
		return "", fmt.Errorf("payment client not configured for COF collection")
	}
	if invoice == nil {
		return "", ErrInvoiceIDRequired
	}
	if inst == nil || strings.TrimSpace(inst.PaymentMethodID) == "" || strings.TrimSpace(inst.CustomerID) == "" {
		return "", fmt.Errorf("saved instrument requires payment_method_id and customer_id")
	}
	route = strings.TrimSpace(route)
	if route == "" {
		route = "flutterwave"
	}
	// Refuse non-flutterwave routes for silent COF in this collector — v4 token path only.
	if !strings.EqualFold(route, "flutterwave") {
		return "", fmt.Errorf("cof collector supports flutterwave v4 only, got %q", route)
	}

	money, err := moneyFromDecimal(invoice.TotalAmount, invoice.Currency)
	if err != nil {
		return nilErr(err)
	}

	extraMap := map[string]any{
		collection.ExtraPaymentMethodID: inst.PaymentMethodID,
		collection.ExtraCustomerID:      inst.CustomerID,
		collection.ExtraRecurring:       "true",
		collection.ExtraAPIVersion:      "v4",
		collection.ExtraInvoiceID:       invoice.GetID(),
		collection.ExtraSubscriptionID:  invoice.SubscriptionID,
		"collection_mode":               "cof",
		collection.ExtraTxRef:           "inv-" + sanitizeRef(invoice.GetID()),
	}
	if inst.Provider != "" {
		extraMap["provider"] = inst.Provider
	}
	extras, _ := structpb.NewStruct(extraMap)

	resp, err := c.pay.InitiatePrompt(ctx, connect.NewRequest(&paymentv1.InitiatePromptRequest{
		Route:  route,
		Amount: money,
		// Profile id as party when available — payment service uses it for tenancy.
		Extra: extras,
	}))
	if err != nil {
		return "", fmt.Errorf("initiate cof prompt: %w", err)
	}
	id := ""
	if data := resp.Msg.GetData(); data != nil {
		id = data.GetId()
	}
	if id == "" {
		return "", fmt.Errorf("initiate cof prompt: empty prompt id")
	}
	return id, nil
}

func nilErr(err error) (string, error) { return "", err }

func (c *paymentCollector) PollPromptStatus(ctx context.Context, promptID string) (commonv1.STATUS, error) {
	if c == nil || c.pay == nil {
		return commonv1.STATUS_UNKNOWN, fmt.Errorf("payment client not configured")
	}
	extras, _ := structpb.NewStruct(map[string]any{"entity_type": "prompt"})
	resp, err := c.pay.Status(ctx, connect.NewRequest(&commonv1.StatusRequest{
		Id:     promptID,
		Extras: extras,
	}))
	if err != nil {
		return commonv1.STATUS_UNKNOWN, err
	}
	return resp.Msg.GetStatus(), nil
}

// --- Instrument resolution ---

type instrumentSource struct {
	profile profilev1connect.ProfileServiceClient
}

// NewInstrumentSource loads COF from subscription.Data then profile checkout clues.
func NewInstrumentSource(profile profilev1connect.ProfileServiceClient) InstrumentSource {
	return &instrumentSource{profile: profile}
}

func (s *instrumentSource) ResolveInstrument(ctx context.Context, sub *models.Subscription) (*SavedInstrument, error) {
	if sub == nil {
		return nil, fmt.Errorf("nil subscription")
	}
	// Prefer instrument pinned on the subscription (product-isolated).
	if inst := instrumentFromData(sub.Data); inst != nil {
		return inst, nil
	}
	// Fall back to profile checkout clues (Link-style).
	if s.profile == nil || strings.TrimSpace(sub.ProfileID) == "" {
		return nil, fmt.Errorf("no saved instrument on subscription and profile client unavailable")
	}
	resp, err := s.profile.GetById(ctx, connect.NewRequest(&profilev1.GetByIdRequest{
		Id: sub.ProfileID,
	}))
	if err != nil {
		return nil, fmt.Errorf("profile get: %w", err)
	}
	props := resp.Msg.GetData().GetProperties()
	if props == nil {
		return nil, fmt.Errorf("profile has no checkout instrument")
	}
	checkout := props.GetFields()["checkout"].GetStructValue()
	if checkout == nil {
		return nil, fmt.Errorf("profile has no checkout clues")
	}
	f := checkout.GetFields()
	pmd := strings.TrimSpace(f["paymentMethodId"].GetStringValue())
	cus := strings.TrimSpace(f["providerCustomerId"].GetStringValue())
	if pmd == "" || cus == "" {
		return nil, fmt.Errorf("profile checkout clues missing paymentMethodId/providerCustomerId")
	}
	return &SavedInstrument{
		PaymentMethodID: pmd,
		CustomerID:      cus,
		Provider:        "flutterwave",
	}, nil
}

func instrumentFromData(d map[string]any) *SavedInstrument {
	if d == nil {
		return nil
	}
	pmd, _ := d[models.SubDataPaymentMethodID].(string)
	cus, _ := d[models.SubDataProviderCustomerID].(string)
	pmd = strings.TrimSpace(pmd)
	cus = strings.TrimSpace(cus)
	if pmd == "" || cus == "" {
		return nil
	}
	prov, _ := d[models.SubDataPaymentProvider].(string)
	if prov == "" {
		prov = "flutterwave"
	}
	return &SavedInstrument{PaymentMethodID: pmd, CustomerID: cus, Provider: prov}
}

func sanitizeRef(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if len(out) > 36 {
		out = out[:36]
	}
	return out
}

// WaitCOF polls until terminal or timeout (bounded for sweeper).
func WaitCOF(
	ctx context.Context,
	col PaymentCollector,
	promptID string,
	timeout time.Duration,
	interval time.Duration,
) (commonv1.STATUS, error) {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	deadline := time.Now().Add(timeout)
	var last commonv1.STATUS
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return last, err
		}
		st, err := col.PollPromptStatus(ctx, promptID)
		if err != nil {
			util.Log(ctx).WithError(err).Debug("cof status poll error")
			time.Sleep(interval)
			continue
		}
		last = st
		if st == commonv1.STATUS_SUCCESSFUL || st == commonv1.STATUS_FAILED {
			return st, nil
		}
		time.Sleep(interval)
	}
	return last, fmt.Errorf("cof charge still pending after %s", timeout)
}
