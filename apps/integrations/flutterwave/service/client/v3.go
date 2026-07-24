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

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// v3BaseURL is the classic Flutterwave API host (test and live share the host;
// key prefix FLWSECK_TEST vs FLWSECK distinguishes environment).
const v3BaseURL = "https://api.flutterwave.com/v3"

// IsV3Credentials reports whether creds are classic secret-key style (FLWSECK_*)
// rather than v4 OAuth client_id/client_secret.
func IsV3Credentials(c *Credentials) bool {
	if c == nil {
		return false
	}
	sec := firstNonEmpty(c.SecretKey, c.ClientSecret)
	return strings.HasPrefix(sec, "FLWSECK_") || strings.HasPrefix(sec, "FLWSECK-")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func (c *Credentials) secretKey() string {
	return firstNonEmpty(c.SecretKey, c.ClientSecret)
}

// PublicKeyValue returns the classic public key when set.
func (c *Credentials) PublicKeyValue() string {
	return firstNonEmpty(c.PublicKey, c.ClientID)
}

// createStandardPaymentV3 implements POST /v3/payments (hosted Standard checkout).
func (c *flutterwaveClient) createStandardPaymentV3(
	ctx context.Context,
	creds *Credentials,
	req *OrchestratorChargeRequest,
) (*Charge, error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "v3_standard_payment")
	var retErr error
	defer func() { done(retErr) }()

	amount := strconv.FormatFloat(req.Amount, 'f', -1, 64)
	phone := ""
	if req.Customer.Phone != nil {
		phone = req.Customer.Phone.CountryCode + req.Customer.Phone.Number
	}
	name := ""
	if req.Customer.Name != nil {
		name = strings.TrimSpace(req.Customer.Name.First + " " + req.Customer.Name.Last)
	}

	paymentOptions := "card, mpesa, mobilemoneyuganda, mobilemoneytanzania, mobilemoneyghana, banktransfer, ussd"
	// Prefer MoMo options when payment method is mobile_money.
	if req.PaymentMethod.Type == "mobile_money" {
		paymentOptions = "mpesa, mobilemoneyuganda, mobilemoneytanzania, mobilemoneyghana, card"
	}

	body := map[string]any{
		"tx_ref":          req.Reference,
		"amount":          amount,
		"currency":        req.Currency,
		"redirect_url":    req.RedirectURL,
		"payment_options": paymentOptions,
		"customer": map[string]any{
			"email":       req.Customer.Email,
			"name":        name,
			"phonenumber": phone,
		},
		"meta": req.Meta,
		"customizations": map[string]any{
			"title": "Payment",
		},
	}

	var out struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Link string `json:"link"`
		} `json:"data"`
	}
	if err := c.doV3JSON(ctx, creds, http.MethodPost, "/payments", body, &out); err != nil {
		retErr = err
		return nil, err
	}
	if !strings.EqualFold(out.Status, "success") || out.Data.Link == "" {
		retErr = fmt.Errorf("flutterwave v3 payments: status=%s message=%s", out.Status, out.Message)
		return nil, retErr
	}

	ch := &Charge{
		ID:          "v3:" + req.Reference,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Reference:   req.Reference,
		Status:      "pending",
		RedirectURL: req.RedirectURL,
		Meta:        stringMapToAny(req.Meta),
		NextAction: map[string]any{
			"type": "redirect_url",
			"redirect_url": map[string]any{
				"url": out.Data.Link,
			},
		},
	}
	return ch, nil
}

// createMoMoChargeV3 implements POST /v3/charges?type=…
func (c *flutterwaveClient) createMoMoChargeV3(
	ctx context.Context,
	creds *Credentials,
	req *OrchestratorChargeRequest,
) (*Charge, error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "v3_momo_charge")
	var retErr error
	defer func() { done(retErr) }()

	mm := req.PaymentMethod.MobileMoney
	if mm == nil {
		retErr = errors.New("mobile_money details required")
		return nil, retErr
	}

	chargeType := momoChargeType(req.Currency, mm.CountryCode)
	phone := mm.CountryCode + mm.PhoneNumber
	amount := strconv.FormatFloat(req.Amount, 'f', -1, 64)
	name := ""
	if req.Customer.Name != nil {
		name = strings.TrimSpace(req.Customer.Name.First + " " + req.Customer.Name.Last)
	}

	body := map[string]any{
		"tx_ref":       req.Reference,
		"amount":       amount,
		"currency":     req.Currency,
		"email":        req.Customer.Email,
		"phone_number": phone,
		"fullname":     name,
		"meta":         req.Meta,
	}
	if mm.Network != "" && chargeType != "mpesa" {
		body["network"] = mm.Network
	}

	var out struct {
		Status  string         `json:"status"`
		Message string         `json:"message"`
		Data    map[string]any `json:"data"`
	}
	path := "/charges?type=" + url.QueryEscape(chargeType)
	if err := c.doV3JSON(ctx, creds, http.MethodPost, path, body, &out); err != nil {
		retErr = err
		return nil, err
	}

	ch := &Charge{
		ID:        "v3:" + req.Reference,
		Amount:    req.Amount,
		Currency:  req.Currency,
		Reference: req.Reference,
		Status:    "pending",
		Meta:      stringMapToAny(req.Meta),
	}
	if out.Data != nil {
		if id, ok := asFloatID(out.Data["id"]); ok {
			ch.ID = strconv.FormatInt(id, 10)
		}
		if st, ok := out.Data["status"].(string); ok {
			ch.Status = st
		}
		// Redirect authorization (UG/GH)
		if meta, ok := out.Data["meta"].(map[string]any); ok {
			if auth, ok := meta["authorization"].(map[string]any); ok {
				if redir, ok := auth["redirect"].(string); ok && redir != "" {
					ch.NextAction = map[string]any{
						"type":         "redirect_url",
						"redirect_url": map[string]any{"url": redir},
					}
				}
			}
		}
		if ch.NextAction == nil {
			ch.NextAction = map[string]any{
				"type": "payment_instruction",
				"payment_instruction": map[string]any{
					"note": "Please authorise this payment on your mobile phone. It may take a few minutes to confirm.",
				},
			}
		}
	}
	return ch, nil
}

func momoChargeType(currency, countryCode string) string {
	switch strings.ToUpper(currency) {
	case "KES":
		return "mpesa"
	case "UGX":
		return "mobile_money_uganda"
	case "TZS":
		return "mobile_money_tanzania"
	case "GHS":
		return "mobile_money_ghana"
	case "RWF":
		return "mobile_money_rwanda"
	default:
		if countryCode == "254" {
			return "mpesa"
		}
		return "mpesa"
	}
}

// getChargeV3 verifies via tx_ref or numeric id.
func (c *flutterwaveClient) getChargeV3(
	ctx context.Context,
	creds *Credentials,
	chargeID string,
) (*Charge, error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "v3_verify")
	var retErr error
	defer func() { done(retErr) }()

	var out struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			ID       int64   `json:"id"`
			TxRef    string  `json:"tx_ref"`
			FlwRef   string  `json:"flw_ref"`
			Amount   float64 `json:"amount"`
			Currency string  `json:"currency"`
			Status   string  `json:"status"`
			Meta     any     `json:"meta"`
		} `json:"data"`
	}

	var path string
	if strings.HasPrefix(chargeID, "v3:") {
		txRef := strings.TrimPrefix(chargeID, "v3:")
		path = "/transactions/verify_by_reference?tx_ref=" + url.QueryEscape(txRef)
	} else if id, err := strconv.ParseInt(chargeID, 10, 64); err == nil {
		path = "/transactions/" + strconv.FormatInt(id, 10) + "/verify"
	} else {
		path = "/transactions/verify_by_reference?tx_ref=" + url.QueryEscape(chargeID)
	}

	if err := c.doV3JSON(ctx, creds, http.MethodGet, path, nil, &out); err != nil {
		retErr = err
		return nil, err
	}

	status := strings.ToLower(out.Data.Status)
	switch status {
	case "successful", "success":
		status = "succeeded"
	case "failed":
		status = "failed"
	default:
		if status == "" {
			status = "pending"
		}
	}

	meta := map[string]any{}
	switch m := out.Data.Meta.(type) {
	case map[string]any:
		meta = m
	}

	return &Charge{
		ID:        strconv.FormatInt(out.Data.ID, 10),
		Amount:    out.Data.Amount,
		Currency:  out.Data.Currency,
		Reference: out.Data.TxRef,
		Status:    status,
		Meta:      meta,
	}, nil
}

// createTransferV3 implements POST /v3/transfers.
func (c *flutterwaveClient) createTransferV3(
	ctx context.Context,
	creds *Credentials,
	req map[string]any,
) (*Transfer, error) {
	ctx, done := c.metrics.ObserveProviderCall(ctx, "v3_transfer")
	var retErr error
	defer func() { done(retErr) }()

	// Map our internal transfer map → v3 body when possible.
	v3Body := req
	if pi, ok := req["payment_instruction"].(map[string]any); ok {
		// Rebuild simplified v3 transfer if orchestrator-shaped.
		ref, _ := req["reference"].(string)
		narr, _ := req["narration"].(string)
		currency, _ := pi["destination_currency"].(string)
		amountVal := 0.0
		if am, ok := pi["amount"].(map[string]any); ok {
			switch v := am["value"].(type) {
			case float64:
				amountVal = v
			case int:
				amountVal = float64(v)
			case int64:
				amountVal = float64(v)
			}
		}
		// Expect bank details in meta or nested recipient — callers of CreateTransfer
		// after recipient creation should pass v3-shaped body from payments.go when v3.
		_ = ref
		_ = narr
		_ = currency
		_ = amountVal
	}

	var out struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			ID        int64   `json:"id"`
			Reference string  `json:"reference"`
			Status    string  `json:"status"`
			Amount    float64 `json:"amount"`
			Currency  string  `json:"currency"`
		} `json:"data"`
	}
	if err := c.doV3JSON(ctx, creds, http.MethodPost, "/transfers", v3Body, &out); err != nil {
		retErr = err
		return nil, err
	}
	if !strings.EqualFold(out.Status, "success") {
		retErr = fmt.Errorf("flutterwave v3 transfer: %s", out.Message)
		return nil, retErr
	}
	return &Transfer{
		ID:        strconv.FormatInt(out.Data.ID, 10),
		Reference: out.Data.Reference,
		Status:    out.Data.Status,
	}, nil
}

func (c *flutterwaveClient) doV3JSON(
	ctx context.Context,
	creds *Credentials,
	method, path string,
	payload any,
	result any,
) error {
	secret := creds.secretKey()
	if secret == "" {
		return errors.New("flutterwave secret key required")
	}

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, v3BaseURL+path, body)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+secret)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("flutterwave v3 http %d: %s", resp.StatusCode, truncate(string(respBody), 512))
	}
	if result == nil || len(respBody) == 0 {
		return nil
	}
	return json.Unmarshal(respBody, result)
}

func stringMapToAny(m map[string]string) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func asFloatID(v any) (int64, bool) {
	switch t := v.(type) {
	case float64:
		return int64(t), true
	case int64:
		return t, true
	case json.Number:
		n, err := t.Int64()
		return n, err == nil
	case string:
		n, err := strconv.ParseInt(t, 10, 64)
		return n, err == nil
	default:
		return 0, false
	}
}
