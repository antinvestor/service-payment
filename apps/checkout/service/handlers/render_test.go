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

package handlers_test

import (
	"bytes"
	"strings"
	"testing"
	"unicode"

	"github.com/antinvestor/service-payments/apps/checkout/service/handlers"
)

// ---- MaskMsisdn tests -------------------------------------------------------

func TestMaskMsisdn(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"+254712345789", "+254 7•• ••789"},
		{"254712345789", "+254 7•• ••789"},
		{"12345", "••••"},
		{"", "••••"},
		{"12345678", "••••"}, // 8 digits — fewer than 9
		{"123456789", "+123 4•• ••789"},
		{"  +44 7911 123456  ", "+447 9•• ••456"},
	}
	for _, tc := range cases {
		got := handlers.MaskMsisdn(tc.input)
		if got != tc.want {
			t.Errorf("MaskMsisdn(%q) = %q; want %q", tc.input, got, tc.want)
		}
	}
}

func TestMaskMsisdnDigitCountProperty(t *testing.T) {
	// The output must never contain more than 7 visible digits.
	inputs := []string{
		"+254712345789",
		"1234567890123",
		"254700000001",
	}
	for _, input := range inputs {
		masked := handlers.MaskMsisdn(input)
		var visible int
		for _, r := range masked {
			if unicode.IsDigit(r) {
				visible++
			}
		}
		if visible > 7 {
			t.Errorf("MaskMsisdn(%q) = %q exposes %d digits; want ≤7", input, masked, visible)
		}
	}
}

// ---- CSRFToken / VerifyCSRF tests -------------------------------------------

func TestCSRFTokenRoundTrip(t *testing.T) {
	secret := []byte("super-secret-key")
	ref := "session-abc-123"

	token := handlers.CSRFToken(secret, ref)
	if token == "" {
		t.Fatal("CSRFToken returned empty string")
	}

	if !handlers.VerifyCSRF(secret, ref, token) {
		t.Error("VerifyCSRF should return true for correct token")
	}
	if handlers.VerifyCSRF(secret, "other-ref", token) {
		t.Error("VerifyCSRF should return false for wrong ref")
	}
	if handlers.VerifyCSRF(secret, ref, "deadbeef"+token) {
		t.Error("VerifyCSRF should return false for tampered token")
	}
	if handlers.VerifyCSRF(secret, ref, "0000000000000000000000000000000000000000000000000000000000000000") {
		t.Error("VerifyCSRF should return false for all-zero token")
	}
}

// ---- Translation tests ------------------------------------------------------

func TestTranslate(t *testing.T) {
	// known en key
	if got := handlers.T("en", "pay_button"); got != "Pay" {
		t.Errorf(`T("en","pay_button") = %q; want "Pay"`, got)
	}
	// known fr key
	if got := handlers.T("fr", "pay_button"); got != "Payer" {
		t.Errorf(`T("fr","pay_button") = %q; want "Payer"`, got)
	}
	// unknown lang falls back to en
	if got := handlers.T("xx", "pay_button"); got != "Pay" {
		t.Errorf(`T("xx","pay_button") = %q; want "Pay" (en fallback)`, got)
	}
	// missing key returns key itself
	if got := handlers.T("en", "no_such_key"); got != "no_such_key" {
		t.Errorf(`T("en","no_such_key") = %q; want "no_such_key"`, got)
	}
}

// ---- Template rendering tests -----------------------------------------------

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func newTestRenderer(t *testing.T) *handlers.Renderer {
	t.Helper()
	r, err := handlers.NewRenderer([]byte("test-secret"))
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	return r
}

func samplePageData(lang, page string) handlers.PageData {
	// Raw msisdn — must NEVER appear in rendered output of any template.
	rawMsisdn := "254712345789"
	return handlers.PageData{
		Lang:          lang,
		SessionRef:    "sess-001",
		MerchantName:  "Acme Shop",
		Description:   "Test payment",
		AmountDisplay: "KES 100.00",
		Variable:      false,
		Currency:      "KES",
		PayerName:     "Alice",
		MaskedPhone:   handlers.MaskMsisdn(rawMsisdn),
		Contacts: []handlers.ContactChoice{
			{ContactID: "c1", Masked: handlers.MaskMsisdn(rawMsisdn)},
		},
		Methods: []handlers.MethodChoice{
			{Key: "mpesa", Name: "M-Pesa", Selected: true},
			{Key: "card", Name: "Card", Selected: false},
		},
		CSRF:          handlers.CSRFToken([]byte("test-secret"), "sess-001"),
		Status:        "pending",
		FailureReason: "",
		ReturnURL:     "https://example.com/return",
		PollURL:       "/c/sess-001/status",
		RedirectURL: func() string {
			if page == "done" {
				return "https://example.com/return"
			}
			return ""
		}(),
	}
}

func TestAllTemplatesRenderInBothLanguages(t *testing.T) {
	r := newTestRenderer(t)
	pages := []string{"pay", "confirm", "done", "gone"}
	langs := []string{"en", "fr"}
	rawMsisdn := "254712345789"

	for _, lang := range langs {
		for _, page := range pages {
			t.Run(lang+"/"+page, func(t *testing.T) {
				data := samplePageData(lang, page)
				var buf bytes.Buffer
				err := r.Render(&buf, page, data)
				if err != nil {
					t.Fatalf("Render(%q) error: %v", page, err)
				}
				out := buf.String()
				if len(out) == 0 {
					t.Fatal("Render produced empty output")
				}
				if strings.Contains(out, rawMsisdn) {
					t.Errorf("Render(%q) leaks raw msisdn %q in output", page, rawMsisdn)
				}
			})
		}
	}
}

// ---- Pay template state tests -----------------------------------------------

func testPayVariableAmount(t *testing.T, r *handlers.Renderer) {
	t.Helper()
	data := samplePageData("en", "pay")
	data.Variable = true
	var buf bytes.Buffer
	if err := r.Render(&buf, "pay", data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(buf.String(), `name="amount"`) {
		t.Error(`expected name="amount" input for variable amount`)
	}
}

func testPayFixedAmount(t *testing.T, r *handlers.Renderer) {
	t.Helper()
	data := samplePageData("en", "pay")
	data.Variable = false
	var buf bytes.Buffer
	if err := r.Render(&buf, "pay", data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, `name="amount"`) {
		t.Error("fixed amount should NOT show amount input")
	}
	if !strings.Contains(out, "amount-display") {
		t.Error("fixed amount should show amount-display div")
	}
}

func testPayRecognizedPayer(t *testing.T, r *handlers.Renderer) {
	t.Helper()
	data := samplePageData("en", "pay")
	data.PayerName = "Alice"
	data.MaskedPhone = handlers.MaskMsisdn("254712345789")
	var buf bytes.Buffer
	if err := r.Render(&buf, "pay", data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := buf.String()
	// html/template escapes '+' as '&#43;', so search for digit substrings.
	if !strings.Contains(out, "789") || !strings.Contains(out, "254") {
		t.Errorf("expected masked phone digits in output; got snippet: %q", truncate(out, 500))
	}
	if !strings.Contains(out, "payer-info") {
		t.Error("expected payer-info section for recognized payer")
	}
	if strings.Contains(out, `name="phone"`) {
		t.Error("recognized payer should NOT show phone input")
	}
}

func testPayGuestPayer(t *testing.T, r *handlers.Renderer) {
	t.Helper()
	data := samplePageData("en", "pay")
	data.PayerName = ""
	data.MaskedPhone = ""
	var buf bytes.Buffer
	if err := r.Render(&buf, "pay", data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(buf.String(), `name="phone"`) {
		t.Error("guest payer should show phone input")
	}
}

func testPayFailureBanner(t *testing.T, r *handlers.Renderer) {
	t.Helper()
	data := samplePageData("en", "pay")
	data.FailureReason = "Insufficient funds"
	var buf bytes.Buffer
	if err := r.Render(&buf, "pay", data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "banner error") {
		t.Error("expected banner error class in output")
	}
	if !strings.Contains(out, "Insufficient funds") {
		t.Error("expected failure reason text in output")
	}
}

func testPayCSRF(t *testing.T, r *handlers.Renderer) {
	t.Helper()
	data := samplePageData("en", "pay")
	data.CSRF = "testtoken123"
	var buf bytes.Buffer
	if err := r.Render(&buf, "pay", data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `name="csrf"`) {
		t.Error("expected hidden csrf input")
	}
	if !strings.Contains(out, "testtoken123") {
		t.Error("expected csrf token value in output")
	}
}

func TestPayTemplateStates(t *testing.T) {
	r := newTestRenderer(t)
	t.Run("variable_amount_shows_input", func(t *testing.T) { testPayVariableAmount(t, r) })
	t.Run("fixed_amount_shows_display", func(t *testing.T) { testPayFixedAmount(t, r) })
	t.Run("recognized_payer_shows_masked_phone_not_phone_input", func(t *testing.T) { testPayRecognizedPayer(t, r) })
	t.Run("guest_shows_phone_input", func(t *testing.T) { testPayGuestPayer(t, r) })
	t.Run("failure_reason_renders_banner", func(t *testing.T) { testPayFailureBanner(t, r) })
	t.Run("csrf_hidden_input_present", func(t *testing.T) { testPayCSRF(t, r) })
}
