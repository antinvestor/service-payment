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

package handlers

import (
	"context"
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/checkout/config"
	checkoutv1 "github.com/antinvestor/service-payments/apps/checkout/gen/checkout/v1"
	"github.com/antinvestor/service-payments/apps/checkout/gen/checkout/v1/checkoutv1connect"
	"github.com/antinvestor/service-payments/apps/checkout/service/business"
	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"gorm.io/gorm"
)

// CheckoutServer implements checkoutv1connect.CheckoutServiceHandler.
type CheckoutServer struct {
	business *business.CheckoutBusiness
	cfg      *config.CheckoutConfig
}

// Ensure interface is satisfied at compile time.
var _ checkoutv1connect.CheckoutServiceHandler = (*CheckoutServer)(nil)

// NewCheckoutServer creates a CheckoutServer wired to the given business layer and config.
func NewCheckoutServer(biz *business.CheckoutBusiness, cfg *config.CheckoutConfig) *CheckoutServer {
	return &CheckoutServer{business: biz, cfg: cfg}
}

// ---------------------------------------------------------------------------
// CreateCheckoutSession
// ---------------------------------------------------------------------------

// CreateCheckoutSession maps the proto request to business.CreateSessionInput and creates a session.
func (s *CheckoutServer) CreateCheckoutSession(
	ctx context.Context,
	req *connect.Request[checkoutv1.CreateCheckoutSessionRequest],
) (*connect.Response[checkoutv1.CreateCheckoutSessionResponse], error) {
	msg := req.Msg

	// Map amount
	amtStr := ""
	currency := ""
	if amt := msg.GetAmount(); amt != nil {
		amtStr = business.AmountString(amt)
		currency = amt.GetCurrencyCode()
	}

	// Map amount option
	amtOption := models.AmountOptionFixed
	if msg.GetAmountOption() == checkoutv1.AmountOption_AMOUNT_OPTION_VARIABLE {
		amtOption = models.AmountOptionVariable
	}

	// Map payer
	var payer *business.PayerInput
	if p := msg.GetPayer(); p != nil {
		contacts := make([]business.PayerContactInput, 0, len(p.GetContacts()))
		for _, c := range p.GetContacts() {
			contacts = append(contacts, business.PayerContactInput{
				ContactID: c.GetContactId(),
				Msisdn:    c.GetMsisdn(),
			})
		}
		payer = &business.PayerInput{
			ProfileID:   p.GetProfileId(),
			DisplayName: p.GetDisplayName(),
			Language:    p.GetLanguage(),
			Contacts:    contacts,
		}
	}

	in := business.CreateSessionInput{
		Name:         msg.GetName(),
		Description:  msg.GetDescription(),
		Amount:       amtStr,
		Currency:     currency,
		AmountOption: amtOption,
		OrderRef:     msg.GetOrderRef(),
		Metadata:     msg.GetMetadata(),
		ReturnURL:    msg.GetReturnUrl(),
		Payer:        payer,
		Methods:      msg.GetMethods(),
	}

	session, err := s.business.CreateSession(ctx, in)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&checkoutv1.CreateCheckoutSessionResponse{
		Data: toProtoSession(session, s.cfg.PublicBaseURL),
	}), nil
}

// ---------------------------------------------------------------------------
// GetCheckoutSession
// ---------------------------------------------------------------------------

// GetCheckoutSession retrieves a session by ref, refreshes payment status, and returns the proto.
func (s *CheckoutServer) GetCheckoutSession(
	ctx context.Context,
	req *connect.Request[checkoutv1.GetCheckoutSessionRequest],
) (*connect.Response[checkoutv1.GetCheckoutSessionResponse], error) {
	session, err := s.business.GetSessionByRef(ctx, req.Msg.GetRef())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Best-effort live refresh — ignore errors, return current state.
	if refreshed, refreshErr := s.business.RefreshStatus(ctx, session); refreshErr == nil {
		session = refreshed
	}

	return connect.NewResponse(&checkoutv1.GetCheckoutSessionResponse{
		Data: toProtoSession(session, s.cfg.PublicBaseURL),
	}), nil
}

// ---------------------------------------------------------------------------
// CreateCheckoutLink
// ---------------------------------------------------------------------------

// CreateCheckoutLink creates a reusable checkout link.
func (s *CheckoutServer) CreateCheckoutLink(
	ctx context.Context,
	req *connect.Request[checkoutv1.CreateCheckoutLinkRequest],
) (*connect.Response[checkoutv1.CreateCheckoutLinkResponse], error) {
	msg := req.Msg

	// Map amount
	amtStr := ""
	currency := ""
	if amt := msg.GetAmount(); amt != nil {
		amtStr = business.AmountString(amt)
		currency = amt.GetCurrencyCode()
	}

	// Map amount option
	amtOption := models.AmountOptionFixed
	if msg.GetAmountOption() == checkoutv1.AmountOption_AMOUNT_OPTION_VARIABLE {
		amtOption = models.AmountOptionVariable
	}

	// Parse optional expires_at
	var expiresAt *time.Time
	if raw := msg.GetExpiresAt(); raw != "" {
		t, parseErr := time.Parse(time.RFC3339, raw)
		if parseErr != nil {
			return nil, connect.NewError(
				connect.CodeInvalidArgument,
				errors.New("expires_at must be RFC3339"),
			)
		}
		expiresAt = &t
	}

	in := business.CreateLinkInput{
		Name:         msg.GetName(),
		Description:  msg.GetDescription(),
		Amount:       amtStr,
		Currency:     currency,
		AmountOption: amtOption,
		OrderRef:     msg.GetOrderRef(),
		ReturnURL:    msg.GetReturnUrl(),
		Metadata:     msg.GetMetadata(),
		ExpiresAt:    expiresAt,
	}

	link, err := s.business.CreateLink(ctx, in)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&checkoutv1.CreateCheckoutLinkResponse{
		Data: toProtoLink(link, s.cfg.PublicBaseURL),
	}), nil
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

// toProtoSession converts a models.CheckoutSession to the proto message.
// Internal metadata keys (prefixed with "_") are filtered out.
func toProtoSession(s *models.CheckoutSession, baseURL string) *checkoutv1.CheckoutSession {
	proto := &checkoutv1.CheckoutSession{
		Ref:         s.Ref,
		Name:        s.Name,
		Description: s.Description,
		OrderRef:    s.OrderRef,
		ReturnUrl:   s.ReturnURL,
		PromptId:    s.PromptID,
		PaymentId:   s.PaymentID,
		PageUrl:     baseURL + "/c/" + s.Ref,
		ExpiresAt:   s.ExpiresAt.Format(time.RFC3339),
	}

	// Amount
	if s.Amount != "" {
		if money, err := business.MoneyFromAmount(s.Amount, s.Currency); err == nil {
			proto.Amount = money
		}
	}

	// AmountOption
	if s.AmountOption == models.AmountOptionVariable {
		proto.AmountOption = checkoutv1.AmountOption_AMOUNT_OPTION_VARIABLE
	} else {
		proto.AmountOption = checkoutv1.AmountOption_AMOUNT_OPTION_FIXED_UNSPECIFIED
	}

	// Status
	proto.Status = sessionStatusToProto(s.Status)

	// Metadata — filter internal underscore-prefixed keys.
	if len(s.Metadata) > 0 {
		m := make(map[string]string)
		for k, v := range s.Metadata {
			if strings.HasPrefix(k, "_") {
				continue
			}
			if str, ok := v.(string); ok {
				m[k] = str
			}
		}
		if len(m) > 0 {
			proto.Metadata = m
		}
	}

	// Payer — only when there is profile ID or non-empty prefill data.
	if s.PayerProfileID != "" || len(s.Prefill) > 0 {
		proto.Payer = buildPayerPrefill(s)
	}

	return proto
}

// toProtoLink converts a models.CheckoutLink to the proto message.
func toProtoLink(l *models.CheckoutLink, baseURL string) *checkoutv1.CheckoutLink {
	proto := &checkoutv1.CheckoutLink{
		Ref:         l.Ref,
		Name:        l.Name,
		Description: l.Description,
		OrderRef:    l.OrderRef,
		ReturnUrl:   l.ReturnURL,
		Active:      l.Active,
		PageUrl:     baseURL + "/l/" + l.Ref,
	}

	// Amount
	if l.Amount != "" {
		if money, err := business.MoneyFromAmount(l.Amount, l.Currency); err == nil {
			proto.Amount = money
		}
	}

	// AmountOption
	if l.AmountOption == models.AmountOptionVariable {
		proto.AmountOption = checkoutv1.AmountOption_AMOUNT_OPTION_VARIABLE
	} else {
		proto.AmountOption = checkoutv1.AmountOption_AMOUNT_OPTION_FIXED_UNSPECIFIED
	}

	// ExpiresAt
	if l.ExpiresAt != nil {
		proto.ExpiresAt = l.ExpiresAt.Format(time.RFC3339)
	}

	// Metadata — filter internal keys.
	if len(l.Metadata) > 0 {
		m := make(map[string]string)
		for k, v := range l.Metadata {
			if strings.HasPrefix(k, "_") {
				continue
			}
			if str, ok := v.(string); ok {
				m[k] = str
			}
		}
		if len(m) > 0 {
			proto.Metadata = m
		}
	}

	return proto
}

// sessionStatusToProto maps the model status string to the proto enum.
func sessionStatusToProto(status string) checkoutv1.SessionStatus {
	switch status {
	case models.SessionStatusProcessing:
		return checkoutv1.SessionStatus_SESSION_STATUS_PROCESSING
	case models.SessionStatusCompleted:
		return checkoutv1.SessionStatus_SESSION_STATUS_COMPLETED
	case models.SessionStatusFailed:
		return checkoutv1.SessionStatus_SESSION_STATUS_FAILED
	case models.SessionStatusExpired:
		return checkoutv1.SessionStatus_SESSION_STATUS_EXPIRED
	default:
		// pending and any unknown value → PENDING_UNSPECIFIED
		return checkoutv1.SessionStatus_SESSION_STATUS_PENDING_UNSPECIFIED
	}
}

// buildPayerPrefill reconstructs a PayerPrefill proto from the session's Prefill JSONMap.
func buildPayerPrefill(s *models.CheckoutSession) *checkoutv1.PayerPrefill {
	payer := &checkoutv1.PayerPrefill{
		ProfileId: s.PayerProfileID,
	}

	if s.Prefill == nil {
		return payer
	}

	if dn, ok := s.Prefill["displayName"].(string); ok {
		payer.DisplayName = dn
	}
	if lang, ok := s.Prefill["language"].(string); ok {
		payer.Language = lang
	}

	// Contacts
	if contactsRaw, ok := s.Prefill["contacts"].([]any); ok {
		for _, raw := range contactsRaw {
			c, isMap := raw.(map[string]any)
			if !isMap {
				continue
			}
			pc := &checkoutv1.PayerContact{}
			if cid, hasCID := c["contactId"].(string); hasCID {
				pc.ContactId = cid
			}
			// msisdn intentionally omitted — do not leak raw phone numbers to merchants
			payer.Contacts = append(payer.Contacts, pc)
		}
	}

	return payer
}
