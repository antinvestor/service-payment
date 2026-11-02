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

// Package fincratest is an in-process stand-in for the Fincra API, shaped
// on the documented requests, responses and error bodies. Tests run the
// real client, workers and webhook server against it.
package fincratest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/antinvestor/service-payments/apps/integrations/fincra/service/client"
)

// Test credentials the server accepts.
const (
	SecretKey     = "sk_test_secret"
	PublicKey     = "pk_test_public"
	BusinessID    = "64f1c9ef838af8cda0211d9d"
	WebhookSecret = "whsec_test"
	ValidOTP      = "123456"
)

// Server is a stateful fake Fincra API.
type Server struct {
	*httptest.Server

	mu        sync.Mutex
	nextID    int
	checkouts map[string]*client.CheckoutPayment
	charges   map[string]*client.Charge
	payouts   map[string]*client.Payout

	// Providers answers GET /core/banks, keyed by "CURRENCY|destination".
	Providers map[string][]client.Provider
	// FailNext makes the next requests to a path prefix (or an exact
	// "METHOD /path") fail with a status.
	FailNext map[string][]int
	// LoseResponse makes the next requests to a path prefix succeed at
	// Fincra but answer with this status, as when a response is lost.
	LoseResponse map[string][]int
	// PayoutError makes payout creation fail with this Fincra errorType.
	PayoutError string

	// Last requests seen, for assertions.
	LastCheckout  *client.CheckoutRequest
	LastCharge    *client.ChargeRequest
	LastPayout    *client.PayoutRequest
	LastAuthorize *client.AuthorizeRequest
	Calls         []string
}

// New starts a fake Fincra API for the test.
func New(t *testing.T) *Server {
	t.Helper()
	s := &Server{
		checkouts: map[string]*client.CheckoutPayment{},
		charges:   map[string]*client.Charge{},
		payouts:   map[string]*client.Payout{},
		Providers: map[string][]client.Provider{},
		FailNext:  map[string][]int{},

		LoseResponse: map[string][]int{},
	}
	s.Server = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.Close)
	return s
}

// Credentials are client credentials pointing at this server.
func (s *Server) Credentials() *client.Credentials {
	return &client.Credentials{
		SecretKey:     SecretKey,
		PublicKey:     PublicKey,
		BusinessID:    BusinessID,
		WebhookSecret: WebhookSecret,
		Environment:   "sandbox",
		BaseURL:       s.URL,
	}
}

// Checkout returns a stored checkout by our reference.
func (s *Server) Checkout(ref string) *client.CheckoutPayment {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.checkouts[ref]
}

// Charge returns a stored charge by our reference.
func (s *Server) Charge(ref string) *client.Charge {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.charges[ref]
}

// Payout returns a stored payout by our reference.
func (s *Server) Payout(ref string) *client.Payout {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.payouts[ref]
}

// PutCheckout stores a checkout as if Fincra had it.
func (s *Server) PutCheckout(cp *client.CheckoutPayment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkouts[cp.MerchantReference] = cp
}

// PutCharge stores a charge as if Fincra had it.
func (s *Server) PutCharge(ch *client.Charge) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.charges[ch.Reference] = ch
}

// PutPayout stores a payout as if Fincra had it.
func (s *Server) PutPayout(p *client.Payout) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.payouts[p.CustomerReference] = p
}

// SetCheckout changes a stored checkout's outcome.
func (s *Server) SetCheckout(ref, status, variance string, received float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := s.checkouts[ref]
	cp.Status, cp.VarianceType, cp.AmountReceived = status, variance, client.FlexFloat(received)
}

// SetPayoutStatus changes a stored payout's status.
func (s *Server) SetPayoutStatus(ref, status, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.payouts[ref]
	p.Status, p.Reason = status, reason
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Calls = append(s.Calls, r.Method+" "+r.URL.Path)
	for prefix, codes := range s.FailNext {
		if len(codes) > 0 && matches(prefix, r) {
			s.FailNext[prefix] = codes[1:]
			writeError(w, codes[0], "INTERNAL_SERVER_ERROR", "Error occured during operation.")
			return
		}
	}

	for prefix, codes := range s.LoseResponse {
		if len(codes) > 0 && r.URL.Path == prefix {
			s.LoseResponse[prefix] = codes[1:]
			s.route(httptest.NewRecorder(), r)
			writeError(w, codes[0], "SERVICE_UNAVAILABLE", "Error occurred during operation. Please try again later.")
			return
		}
	}
	s.route(w, r)
}

func (s *Server) route(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("api-key") != SecretKey {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"message": "Invalid authentication credentials"})
		return
	}

	path := r.URL.Path
	switch {
	case r.Method == http.MethodPost && path == "/checkout/payments":
		s.createCheckout(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/checkout/payments/merchant-reference/"):
		s.lookup(w, r, s.checkouts[strings.TrimPrefix(path, "/checkout/payments/merchant-reference/")] != nil,
			func() any { return s.checkouts[strings.TrimPrefix(path, "/checkout/payments/merchant-reference/")] })
	case r.Method == http.MethodPost && path == "/checkout/charges":
		s.createCharge(w, r)
	case r.Method == http.MethodPost && strings.HasPrefix(path, "/checkout/charges/") && strings.HasSuffix(path, "/authorize"):
		s.authorize(w, r, strings.TrimSuffix(strings.TrimPrefix(path, "/checkout/charges/"), "/authorize"))
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/checkout/charges/merchant-reference/"):
		ref := strings.TrimPrefix(path, "/checkout/charges/merchant-reference/")
		s.lookup(w, r, s.charges[ref] != nil, func() any { return s.charges[ref] })
	case r.Method == http.MethodPost && path == "/disbursements/payouts":
		s.createPayout(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/disbursements/payouts/customer-reference/"):
		ref := strings.TrimPrefix(path, "/disbursements/payouts/customer-reference/")
		s.lookup(w, r, s.payouts[ref] != nil, func() any { return s.payouts[ref] })
	case r.Method == http.MethodGet && path == "/core/banks":
		q := r.URL.Query()
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true, "message": "Banks fetched successfully",
			"data": s.Providers[q.Get("currency")+"|"+q.Get("paymentDestination")],
		})
	default:
		writeError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "route not found")
	}
}

func (s *Server) id(prefix string) string {
	s.nextID++
	return fmt.Sprintf("%s%d", prefix, 1000+s.nextID)
}

func (s *Server) requireBusiness(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("x-business-id") != BusinessID {
		writeError(w, http.StatusForbidden, "ACCESS_DENIED", "Access denied to requested resource")
		return false
	}
	return true
}

func (s *Server) lookup(w http.ResponseWriter, r *http.Request, found bool, record func() any) {
	if !s.requireBusiness(w, r) {
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": true, "message": "fetched", "data": record()})
}

func (s *Server) createCheckout(w http.ResponseWriter, r *http.Request) {
	if !s.requireBusiness(w, r) {
		return
	}
	if r.Header.Get("x-pub-key") != PublicKey {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"status": false, "message": "Invalid public key"})
		return
	}
	var req client.CheckoutRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "Validation failed")
		return
	}
	s.LastCheckout = &req
	if _, dup := s.checkouts[req.Reference]; dup {
		writeError(w, http.StatusUnprocessableEntity, "DUPLICATE_REFERENCE", "Duplicate reference")
		return
	}
	payCode := s.id("fcr-p-")
	meta := map[string]any{}
	for k, v := range req.Metadata {
		meta[k] = v
	}
	s.checkouts[req.Reference] = &client.CheckoutPayment{
		ID: client.FlexString(s.id("")), Amount: client.FlexFloat(req.Amount), AmountExpected: client.FlexFloat(req.Amount),
		Currency: req.Currency, Reference: payCode, MerchantReference: req.Reference, Status: client.StatusInitiated,
		FeeBearer: req.FeeBearer, Metadata: meta, Customer: req.Customer,
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": true, "message": "Hosted link generated", "data": map[string]any{
		"link": "https://sandbox-checkout.fincra.com/pay/" + payCode, "reference": req.Reference, "payCode": payCode,
	}})
}

func (s *Server) createCharge(w http.ResponseWriter, r *http.Request) {
	if !s.requireBusiness(w, r) {
		return
	}
	var req client.ChargeRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "Validation failed")
		return
	}
	s.LastCharge = &req
	if _, dup := s.charges[req.Reference]; dup {
		writeError(w, http.StatusUnprocessableEntity, "DUPLICATE_REFERENCE", "Duplicate reference")
		return
	}
	ch := &client.Charge{
		ID: client.FlexString(s.id("")), Amount: client.FlexFloat(req.Amount), AmountExpected: client.FlexFloat(req.Amount),
		Currency: req.Currency, Status: client.StatusPending, Reference: req.Reference, Type: req.Type,
		Customer: req.Customer, ChargeReference: s.id("fcr-ch-"),
	}
	switch req.Type {
	case client.ChargeTypeMobileMoney:
		if req.Phone == "" || req.Operator == "" {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "phone and operator are required")
			return
		}
		if req.Currency == "GHS" {
			ch.AuthModel, ch.Authorization.Mode = client.AuthModelOTP, client.AuthModelOTP
			ch.Message = "Enter the OTP sent to your phone"
		} else {
			ch.Message = "<p>You will receive a prompt on mobile number <b>" + req.Phone +
				"</b>. <br/> Enter your PIN to authorize your payment.</p>"
		}
	case client.ChargeTypeBankTransfer:
		ch.VirtualAccount = &client.VirtualAccount{
			ID: "va-1", AccountNumber: "9600414830", AccountName: "Fincra Checkout", BankCode: "109",
			BankName: "Test Bank", ExpiresAt: "2030-01-01T16:17:56.180Z",
		}
	default:
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "unsupported type")
		return
	}
	s.charges[req.Reference] = ch
	writeJSON(w, http.StatusOK, map[string]any{"status": true, "message": "Charge created", "data": ch})
}

func (s *Server) authorize(w http.ResponseWriter, r *http.Request, id string) {
	if !s.requireBusiness(w, r) {
		return
	}
	var req client.AuthorizeRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	s.LastAuthorize = &req
	for _, ch := range s.charges {
		if ch.ID.String() != id {
			continue
		}
		if req.OTP != ValidOTP {
			writeJSON(w, http.StatusBadRequest, map[string]any{"status": false, "message": "Invalid OTP supplied"})
			return
		}
		ch.Status, ch.Message = client.StatusSuccess, "Mobile money charged successfully"
		writeJSON(w, http.StatusOK, map[string]any{"status": true, "message": "Charge authorized", "data": ch})
		return
	}
	writeError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Charge not found")
}

func (s *Server) createPayout(w http.ResponseWriter, r *http.Request) {
	if !s.requireBusiness(w, r) {
		return
	}
	var req client.PayoutRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "Validation failed")
		return
	}
	s.LastPayout = &req
	if req.Business != BusinessID {
		writeError(w, http.StatusForbidden, "OPERATION_FORBIDDEN", "Operation forbidden")
		return
	}
	if s.PayoutError != "" {
		writeError(w, http.StatusUnprocessableEntity, s.PayoutError, "You don't have enough money in your wallet to make this payout")
		return
	}
	if _, dup := s.payouts[req.CustomerReference]; dup {
		writeError(w, http.StatusUnprocessableEntity, client.ErrTypeDuplicateReference,
			"Cannot continue, Duplicate Customer Reference Passed")
		return
	}
	amount, _ := req.Amount.Float64()
	p := &client.Payout{
		ID: client.FlexString(s.id("")), AmountSent: client.FlexFloat(amount), AmountReceived: client.FlexFloat(amount),
		SourceCurrency: req.SourceCurrency, DestinationCurrency: req.DestinationCurrency, Status: client.StatusProcessing,
		Reference: s.id("ref"), CustomerReference: req.CustomerReference, PaymentDestination: req.PaymentDestination,
		Recipient: client.PayoutRecipient{AccountNumber: req.Beneficiary.AccountNumber},
	}
	s.payouts[req.CustomerReference] = p
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Payout initiated successfully.", "data": map[string]any{
		"id": p.ID, "reference": p.Reference, "customerReference": p.CustomerReference, "status": p.Status,
		"isDocumentRequired": false, "documentsRequired": []string{},
	}})
}

// matches reports whether key ("METHOD /path" exactly, or a path prefix)
// names the request.
func matches(key string, r *http.Request) bool {
	if method, path, ok := strings.Cut(key, " "); ok {
		return r.Method == method && r.URL.Path == path
	}
	return strings.HasPrefix(r.URL.Path, key)
}

func writeError(w http.ResponseWriter, status int, errorType, message string) {
	writeJSON(w, status, map[string]any{"success": false, "error": message, "errorType": errorType})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
