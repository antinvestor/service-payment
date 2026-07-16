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
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/antinvestor/service-payments/apps/checkout/service/business"
	"github.com/antinvestor/service-payments/apps/checkout/service/models"
	"github.com/pitabwire/util"
)

// internalCreateSessionRequest is the JSON body for POST /internal/v1/sessions.
// Used by product services (e.g. opportunities matching) that open pay.* for
// the payer. Authenticated with X-Checkout-Internal-Token — not Connect RPC.
type internalCreateSessionRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Amount      string            `json:"amount"`   // major units decimal, e.g. "10.00"
	Currency    string            `json:"currency"` // ISO 4217
	OrderRef    string            `json:"order_ref"`
	ReturnURL   string            `json:"return_url"`
	Methods     []string          `json:"methods"`
	Metadata    map[string]string `json:"metadata"`
	ProfileID   string            `json:"profile_id"`
	DisplayName string            `json:"display_name"`
	Email       string            `json:"email"`
	Phone       string            `json:"phone"`
	Language    string            `json:"language"`
}

// HandleInternalCreateSession creates a hosted checkout session for trusted
// in-cluster callers and returns {ref, page_url} so the SPA can redirect to
// pay.stawi.org/c/{ref} without Flutterwave multipay.
func (s *WebServer) HandleInternalCreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.authorizeInternal(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var in internalCreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Currency) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name_and_currency_required"})
		return
	}
	if strings.TrimSpace(in.Amount) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "amount_required"})
		return
	}

	// Optional product-supplied email; profile EMAIL contacts still win when
	// profile_id is set and applyPayer finds a contact (caller email is only
	// used when the profile has no EMAIL contact).
	meta := in.Metadata
	if meta == nil {
		meta = map[string]string{}
	}

	var payer *business.PayerInput
	if in.ProfileID != "" || in.DisplayName != "" || in.Phone != "" || in.Email != "" {
		payer = &business.PayerInput{
			ProfileID:   in.ProfileID,
			DisplayName: in.DisplayName,
			Language:    in.Language,
			Email:       in.Email,
		}
		if in.Phone != "" {
			payer.Contacts = []business.PayerContactInput{{
				ContactID: in.Phone,
				Msisdn:    in.Phone,
			}}
		}
	}

	session, err := s.business.CreateSession(r.Context(), business.CreateSessionInput{
		Name:         in.Name,
		Description:  in.Description,
		Amount:       in.Amount,
		Currency:     strings.ToUpper(strings.TrimSpace(in.Currency)),
		AmountOption: models.AmountOptionFixed,
		OrderRef:     in.OrderRef,
		ReturnURL:    in.ReturnURL,
		Metadata:     meta,
		Payer:        payer,
		Methods:      in.Methods,
	})
	if err != nil {
		util.Log(r.Context()).WithError(err).Warn("internal create session failed")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	base := strings.TrimRight(s.cfg.PublicBaseURL, "/")
	pageURL := base + "/c/" + session.Ref
	util.Log(r.Context()).
		WithField("session_ref", session.Ref).
		WithField("page_url", pageURL).
		Info("internal checkout session created")
	writeJSON(w, http.StatusOK, map[string]string{
		"ref":      session.Ref,
		"page_url": pageURL,
		"pageUrl":  pageURL,
	})
}

// HandleInternalGetSession returns session status for trusted product services.
//
//	GET /internal/v1/sessions/{ref}
//	GET /internal/v1/sessions?order_ref=chk_…
//
// Auth: X-Checkout-Internal-Token (same as create).
func (s *WebServer) HandleInternalGetSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.authorizeInternal(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	ref := strings.TrimSpace(r.PathValue("ref"))
	orderRef := strings.TrimSpace(r.URL.Query().Get("order_ref"))

	var (
		session *models.CheckoutSession
		err     error
	)
	switch {
	case ref != "":
		session, err = s.business.GetSessionByRef(r.Context(), ref)
	case orderRef != "":
		session, err = s.business.GetSessionByOrderRef(r.Context(), orderRef)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ref_or_order_ref_required"})
		return
	}
	if err != nil {
		if isNotFoundErr(err) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		}
		util.Log(r.Context()).WithError(err).Warn("internal get session failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"ref":       session.Ref,
		"status":    session.Status,
		"prompt_id": session.PromptID,
		"order_ref": session.OrderRef,
	})
}

func (s *WebServer) authorizeInternal(r *http.Request) bool {
	token := strings.TrimSpace(r.Header.Get("X-Checkout-Internal-Token"))
	if token == "" {
		if auth := r.Header.Get("Authorization"); strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			token = strings.TrimSpace(auth[7:])
		}
	}
	expected := s.cfg.ResolvedInternalToken()
	if expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}
