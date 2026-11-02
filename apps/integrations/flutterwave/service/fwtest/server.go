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

// Package fwtest is a stateful fake of the Flutterwave v4 API for tests:
// OAuth, orchestrator and token charges, charge authorization, transfer
// recipients and transfers, with idempotency keys honoured the way
// Flutterwave documents them and faults that can be injected per request.
package fwtest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/client"
)

// Server is the fake Flutterwave API.
type Server struct {
	*httptest.Server

	mu sync.Mutex
	// ChargeStatus is the status a new charge gets (default "pending").
	ChargeStatus string
	// ChargedAmount, when set, is what a new charge reports as charged
	// instead of the amount asked for (to model a mischarge).
	ChargedAmount float64
	// NextAction is a new charge's next_action (default requires_pin).
	NextAction map[string]any
	// ValidOTP is the OTP that authorizes a charge (default "123456").
	ValidOTP string
	// TransferStatus is the status a new transfer gets (default "NEW").
	TransferStatus string

	// failNext answers the next n requests matching "METHOD /path" (exact
	// path) or "METHOD /prefix*" with an HTTP status, without acting.
	failNext map[string]*fault
	// LoseResponse acts on the next matching request but answers 502, as a
	// response lost on the way back would look.
	LoseResponse map[string]bool

	Charges   map[string]*client.Charge   // by id
	Transfers map[string]*client.Transfer // by id
	idem      map[string][]byte           // idempotency key → first answer
	Calls     map[string]int              // "METHOD /path" → count
	Tokens    int
	seq       int
}

// New starts a fake Flutterwave API, closed with the test.
func New(t *testing.T) *Server {
	t.Helper()
	s := &Server{
		ChargeStatus:   "pending",
		ValidOTP:       "123456",
		TransferStatus: "NEW",
		failNext:       map[string]*fault{},
		LoseResponse:   map[string]bool{},
		Charges:        map[string]*client.Charge{},
		Transfers:      map[string]*client.Transfer{},
		idem:           map[string][]byte{},
		Calls:          map[string]int{},
	}
	s.Server = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.Close)
	return s
}

// Credentials are v4 OAuth credentials pointed at the fake.
func (s *Server) Credentials() *client.Credentials {
	return &client.Credentials{
		ClientID:      "client-id",
		ClientSecret:  "client-secret",
		Environment:   "sandbox",
		APIBaseURL:    s.URL,
		OAuthTokenURL: s.URL + "/token",
		WebhookSecret: "whsec",
	}
}

// Count is how many requests matched "METHOD /path".
func (s *Server) Count(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Calls[key]
}

// ChargeCount is how many charges exist.
func (s *Server) ChargeCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.Charges)
}

// TransferCount is how many transfers exist.
func (s *Server) TransferCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.Transfers)
}

// SetCharge changes a stored charge (e.g. to settle it).
func (s *Server) SetCharge(id string, fn func(*client.Charge)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(s.Charges[id])
}

// SetTransfer changes a stored transfer.
func (s *Server) SetTransfer(id string, fn func(*client.Transfer)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(s.Transfers[id])
}

type fault struct {
	status, times int
}

// Fail answers the next request matching key with status.
func (s *Server) Fail(key string, status int) {
	s.FailN(key, status, 1)
}

// FailN answers the next n requests matching key with status.
func (s *Server) FailN(key string, status, n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failNext[key] = &fault{status: status, times: n}
}

// Lose performs the next request matching key and answers 502.
func (s *Server) Lose(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LoseResponse[key] = true
}

func (s *Server) take(method, path string) (int, bool) {
	for k, f := range s.failNext {
		if matches(k, method, path) {
			f.times--
			if f.times <= 0 {
				delete(s.failNext, k)
			}
			return f.status, true
		}
	}
	return 0, false
}

func (s *Server) takeLose(method, path string) bool {
	for k := range s.LoseResponse {
		if matches(k, method, path) {
			delete(s.LoseResponse, k)
			return true
		}
	}
	return false
}

func matches(key, method, path string) bool {
	m, p, _ := strings.Cut(key, " ")
	if m != method {
		return false
	}
	if strings.HasSuffix(p, "*") {
		return strings.HasPrefix(path, strings.TrimSuffix(p, "*"))
	}
	return p == path
}

//nolint:gocognit,funlen // one fake API router.
func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := r.Method + " " + r.URL.Path
	s.Calls[key]++

	if status, ok := s.take(r.Method, r.URL.Path); ok {
		http.Error(w, `{"status":"failed","message":"injected"}`, status)
		return
	}

	if r.URL.Path == "/token" {
		s.Tokens++
		writeJSON(w, http.StatusOK, map[string]any{"access_token": fmt.Sprintf("tok-%d", s.Tokens), "expires_in": 600})
		return
	}
	if r.Header.Get("Authorization") == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"status": "failed"})
		return
	}

	idemKey := r.Header.Get("X-Idempotency-Key")
	if r.Method != http.MethodGet && idemKey != "" {
		if cached, ok := s.idem[idemKey]; ok {
			w.Header().Set("X-Idempotency-Cache-Hit", "true")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(cached)
			return
		}
	}

	body, _ := io.ReadAll(r.Body)
	status, answer := s.route(r, body)
	raw, _ := json.Marshal(answer)
	if r.Method != http.MethodGet && idemKey != "" && status < 300 {
		s.idem[idemKey] = raw
	}
	if s.takeLose(r.Method, r.URL.Path) {
		http.Error(w, "upstream lost", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}

func envelope(data any) map[string]any {
	return map[string]any{"status": "success", "message": "ok", "data": data}
}

func failure(message string) map[string]any {
	return map[string]any{"status": "failed", "message": message, "error": map[string]any{
		"type": "REQUEST_NOT_VALID", "code": "10400", "message": message,
	}}
}

//nolint:gocognit,funlen,cyclop // one fake API router.
func (s *Server) route(r *http.Request, body []byte) (int, any) {
	path := r.URL.Path
	switch {
	case r.Method == http.MethodPost && (path == "/orchestration/direct-charges" || path == "/charges"):
		var req struct {
			Amount    float64           `json:"amount"`
			Currency  string            `json:"currency"`
			Reference string            `json:"reference"`
			Meta      map[string]string `json:"meta"`
		}
		_ = json.Unmarshal(body, &req)
		for _, ch := range s.Charges {
			if ch.Reference == req.Reference {
				return http.StatusConflict, failure("duplicate reference")
			}
		}
		s.seq++
		meta := map[string]any{}
		for k, v := range req.Meta {
			meta[k] = v
		}
		amount := req.Amount
		if s.ChargedAmount != 0 {
			amount = s.ChargedAmount
		}
		next := s.NextAction
		if next == nil && s.ChargeStatus == "pending" {
			next = map[string]any{"type": "requires_pin"}
		}
		ch := &client.Charge{
			ID:         fmt.Sprintf("chg_%d", s.seq),
			Amount:     amount,
			Currency:   req.Currency,
			Reference:  req.Reference,
			Status:     s.ChargeStatus,
			NextAction: next,
			Meta:       meta,
		}
		s.Charges[ch.ID] = ch
		return http.StatusCreated, envelope(ch)

	case r.Method == http.MethodPut && strings.HasPrefix(path, "/charges/"):
		ch, ok := s.Charges[strings.TrimPrefix(path, "/charges/")]
		if !ok {
			return http.StatusNotFound, failure("charge not found")
		}
		var req client.UpdateChargeRequest
		_ = json.Unmarshal(body, &req)
		switch req.Authorization.Type {
		case "pin":
			ch.NextAction = map[string]any{"type": "requires_otp"}
		case "otp":
			if req.Authorization.OTP == nil || req.Authorization.OTP.Code != s.ValidOTP {
				return http.StatusBadRequest, failure("invalid otp")
			}
			ch.Status = "succeeded"
			ch.NextAction = nil
		case "avs":
			ch.NextAction = map[string]any{"type": "requires_otp"}
		default:
			return http.StatusBadRequest, failure("unknown authorization")
		}
		return http.StatusOK, envelope(ch)

	case r.Method == http.MethodGet && path == "/charges":
		ref := r.URL.Query().Get("reference")
		out := []*client.Charge{}
		for _, ch := range s.Charges {
			if ch.Reference == ref {
				out = append(out, ch)
			}
		}
		return http.StatusOK, envelope(out)

	case r.Method == http.MethodGet && strings.HasPrefix(path, "/charges/"):
		ch, ok := s.Charges[strings.TrimPrefix(path, "/charges/")]
		if !ok {
			return http.StatusNotFound, failure("charge not found")
		}
		return http.StatusOK, envelope(ch)

	case r.Method == http.MethodPost && path == "/transfers/recipients":
		s.seq++
		return http.StatusCreated, envelope(map[string]any{"id": fmt.Sprintf("rcb_%d", s.seq)})

	case r.Method == http.MethodPost && path == "/transfers":
		var req map[string]any
		_ = json.Unmarshal(body, &req)
		ref, _ := req["reference"].(string)
		for _, tr := range s.Transfers {
			if tr.Reference == ref {
				return http.StatusConflict, failure("duplicate reference")
			}
		}
		s.seq++
		meta := map[string]any{}
		if m, ok := req["meta"].(map[string]any); ok {
			meta = m
		}
		amount := map[string]any{}
		currency := ""
		if pi, ok := req["payment_instruction"].(map[string]any); ok {
			if a, ok := pi["amount"].(map[string]any); ok {
				amount = a
			}
			currency, _ = pi["destination_currency"].(string)
		}
		tr := &client.Transfer{
			ID:                  fmt.Sprintf("trf_%d", s.seq),
			Reference:           ref,
			Status:              s.TransferStatus,
			DestinationCurrency: currency,
			Amount:              amount,
			Meta:                meta,
		}
		s.Transfers[tr.ID] = tr
		return http.StatusCreated, envelope(tr)

	case r.Method == http.MethodGet && strings.HasPrefix(path, "/transfers/"):
		tr, ok := s.Transfers[strings.TrimPrefix(path, "/transfers/")]
		if !ok {
			return http.StatusNotFound, failure("transfer not found")
		}
		return http.StatusOK, envelope(tr)
	}
	return http.StatusNotFound, failure("no route " + r.Method + " " + path)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
