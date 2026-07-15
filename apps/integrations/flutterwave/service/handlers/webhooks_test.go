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
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	commonv1 "buf.build/gen/go/antinvestor/common/protocolbuffers/go/common/v1"
	"buf.build/gen/go/antinvestor/payment/connectrpc/go/v1/paymentv1connect"
	"connectrpc.com/connect"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/config"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/client"
	"github.com/antinvestor/service-payments/apps/integrations/flutterwave/service/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubPaymentClient struct {
	paymentv1connect.PaymentServiceClient
	last *commonv1.StatusUpdateRequest
}

func (s *stubPaymentClient) StatusUpdate(
	_ context.Context,
	req *connect.Request[commonv1.StatusUpdateRequest],
) (*connect.Response[commonv1.StatusUpdateResponse], error) {
	s.last = req.Msg
	return connect.NewResponse(&commonv1.StatusUpdateResponse{}), nil
}

type stubFW struct{}

func (f *stubFW) CreateOrchestratorCharge(context.Context, *client.Credentials, *client.OrchestratorChargeRequest) (*client.Charge, error) {
	return nil, nil
}
func (f *stubFW) GetCharge(context.Context, *client.Credentials, string) (*client.Charge, error) {
	return nil, nil
}
func (f *stubFW) CreateTransferRecipient(context.Context, *client.Credentials, *client.TransferRecipientRequest) (string, error) {
	return "", nil
}
func (f *stubFW) CreateTransfer(context.Context, *client.Credentials, map[string]any) (*client.Transfer, error) {
	return nil, nil
}
func (f *stubFW) CreateDirectTransfer(context.Context, *client.Credentials, *client.DirectTransferRequest) (*client.Transfer, error) {
	return nil, nil
}
func (f *stubFW) GetTransfer(context.Context, *client.Credentials, string) (*client.Transfer, error) {
	return nil, nil
}
func (f *stubFW) VerifyWebhookSignature(rawBody []byte, signatureHeader, secretHash string) bool {
	mac := hmac.New(sha256.New, []byte(secretHash))
	_, _ = mac.Write(rawBody)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return expected == signatureHeader
}

var _ client.FlutterwaveClient = (*stubFW)(nil)

func sign(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func TestWebhook_ChargeCompleted(t *testing.T) {
	pay := &stubPaymentClient{}
	fw := &stubFW{}
	cfg := &config.FlutterwaveConfig{WebhookSecret: "sekrit"}
	srv := handlers.NewFlutterwaveWebhookServer(pay, fw, cfg)

	body, _ := json.Marshal(map[string]any{
		"webhook_id": "wbk_1",
		"timestamp":  1,
		"type":       "charge.completed",
		"data": map[string]any{
			"id":        "chg_abc",
			"reference": "prompt-xyz",
			"status":    "succeeded",
			"meta":      map[string]any{"prompt_id": "xyz"},
			"amount":    100.0,
			"currency":  "KES",
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/webhook/flutterwave", bytes.NewReader(body))
	req.Header.Set("flutterwave-signature", sign(body, "sekrit"))
	rr := httptest.NewRecorder()
	srv.NewRouterV1().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, pay.last)
	assert.Equal(t, "xyz", pay.last.GetId())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, pay.last.GetStatus())
	assert.Equal(t, "prompt", pay.last.GetExtras().GetFields()["entity_type"].GetStringValue())
}

func TestWebhook_TransferDisburse(t *testing.T) {
	pay := &stubPaymentClient{}
	fw := &stubFW{}
	cfg := &config.FlutterwaveConfig{WebhookSecret: "sekrit"}
	srv := handlers.NewFlutterwaveWebhookServer(pay, fw, cfg)

	body, _ := json.Marshal(map[string]any{
		"webhook_id": "wbk_2",
		"timestamp":  1,
		"type":       "transfer.disburse",
		"data": map[string]any{
			"id":        "trf_1",
			"reference": "pay-payment-xyz",
			"status":    "SUCCESSFUL",
			"meta":      map[string]any{"payment_id": "payment-xyz"},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/webhook/flutterwave", bytes.NewReader(body))
	req.Header.Set("flutterwave-signature", sign(body, "sekrit"))
	rr := httptest.NewRecorder()
	srv.NewRouterV1().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, pay.last)
	assert.Equal(t, "payment-xyz", pay.last.GetId())
	assert.Equal(t, commonv1.STATUS_SUCCESSFUL, pay.last.GetStatus())
}

func TestWebhook_RejectsBadSignature(t *testing.T) {
	pay := &stubPaymentClient{}
	fw := &stubFW{}
	cfg := &config.FlutterwaveConfig{WebhookSecret: "sekrit"}
	srv := handlers.NewFlutterwaveWebhookServer(pay, fw, cfg)

	req := httptest.NewRequest(http.MethodPost, "/webhook/flutterwave", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("flutterwave-signature", "wrong")
	rr := httptest.NewRecorder()
	srv.NewRouterV1().ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
