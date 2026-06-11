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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pitabwire/util"
)

const httpTimeout = 30 * time.Second

type client struct {
	httpClient *http.Client
}

// NewClient creates a new pawaPay Merchant API v2 client.
func NewClient() PawapayClient {
	return &client{
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

// doRequest performs an authenticated JSON request against the pawaPay API
// and unmarshals the response into result. pawaPay returns domain rejections
// (status REJECTED with a failureReason) with HTTP 200; non-2xx responses
// indicate request, authentication or platform errors and are returned as errors.
func (c *client) doRequest(
	ctx context.Context,
	creds *Credentials,
	method, endpoint string,
	payload any,
	result any,
	logType string,
) error {
	logger := util.Log(ctx).WithField("type", logType)
	defer logger.Release()

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal %s payload: %w", logType, err)
		}
		body = bytes.NewReader(raw)
	}

	apiURL := creds.ResolveBaseURL() + endpoint
	httpReq, err := http.NewRequestWithContext(ctx, method, apiURL, body)
	if err != nil {
		return fmt.Errorf("create %s request: %w", logType, err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+creds.APIToken)
	if payload != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("execute %s request: %w", logType, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	logger.WithFields(map[string]any{
		"status":   resp.StatusCode,
		"response": string(respBody),
	}).Debug(logType + " response")

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%s failed with status %d: %s", logType, resp.StatusCode, string(respBody))
	}

	if err = json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("decode %s response: %w", logType, err)
	}

	return nil
}

func (c *client) InitiateDeposit(
	ctx context.Context,
	creds *Credentials,
	req *DepositRequest,
) (*DepositInitiationResponse, error) {
	payload := depositInitiationPayload{
		DepositID:            req.DepositID,
		Amount:               req.Amount,
		Currency:             req.Currency,
		Payer:                NewMMOParty(req.PhoneNumber, req.Provider),
		PreAuthorisationCode: req.PreAuthorisationCode,
		ClientReferenceID:    req.ClientReferenceID,
		CustomerMessage:      req.CustomerMessage,
		Metadata:             BuildMetadata(req.Metadata),
	}

	var resp DepositInitiationResponse
	if err := c.doRequest(ctx, creds, http.MethodPost, "/v2/deposits", payload, &resp, "deposit"); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) GetDeposit(
	ctx context.Context,
	creds *Credentials,
	depositID string,
) (*DepositStatusResult, error) {
	var resp DepositStatusResult
	endpoint := "/v2/deposits/" + url.PathEscape(depositID)
	if err := c.doRequest(ctx, creds, http.MethodGet, endpoint, nil, &resp, "deposit status"); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) ResendDepositCallback(
	ctx context.Context,
	creds *Credentials,
	depositID string,
) (*ManualActionResponse, error) {
	var resp ManualActionResponse
	endpoint := "/v2/deposits/resend-callback/" + url.PathEscape(depositID)
	if err := c.doRequest(ctx, creds, http.MethodPost, endpoint, nil, &resp, "deposit resend callback"); err != nil {
		return nil, err
	}
	return &resp, nil
}

func payoutPayload(req *PayoutRequest) payoutInitiationPayload {
	return payoutInitiationPayload{
		PayoutID:          req.PayoutID,
		Amount:            req.Amount,
		Currency:          req.Currency,
		Recipient:         NewMMOParty(req.PhoneNumber, req.Provider),
		ClientReferenceID: req.ClientReferenceID,
		CustomerMessage:   req.CustomerMessage,
		Metadata:          BuildMetadata(req.Metadata),
	}
}

func (c *client) InitiatePayout(
	ctx context.Context,
	creds *Credentials,
	req *PayoutRequest,
) (*PayoutInitiationResponse, error) {
	var resp PayoutInitiationResponse
	if err := c.doRequest(ctx, creds, http.MethodPost, "/v2/payouts", payoutPayload(req), &resp, "payout"); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) InitiateBulkPayouts(
	ctx context.Context,
	creds *Credentials,
	reqs []*PayoutRequest,
) ([]*PayoutInitiationResponse, error) {
	payloads := make([]payoutInitiationPayload, 0, len(reqs))
	for _, req := range reqs {
		payloads = append(payloads, payoutPayload(req))
	}

	var resp []*PayoutInitiationResponse
	if err := c.doRequest(
		ctx,
		creds,
		http.MethodPost,
		"/v2/payouts/bulk",
		payloads,
		&resp,
		"bulk payouts",
	); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *client) GetPayout(
	ctx context.Context,
	creds *Credentials,
	payoutID string,
) (*PayoutStatusResult, error) {
	var resp PayoutStatusResult
	endpoint := "/v2/payouts/" + url.PathEscape(payoutID)
	if err := c.doRequest(ctx, creds, http.MethodGet, endpoint, nil, &resp, "payout status"); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) ResendPayoutCallback(
	ctx context.Context,
	creds *Credentials,
	payoutID string,
) (*ManualActionResponse, error) {
	var resp ManualActionResponse
	endpoint := "/v2/payouts/resend-callback/" + url.PathEscape(payoutID)
	if err := c.doRequest(ctx, creds, http.MethodPost, endpoint, nil, &resp, "payout resend callback"); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) CancelEnqueuedPayout(
	ctx context.Context,
	creds *Credentials,
	payoutID string,
) (*ManualActionResponse, error) {
	var resp ManualActionResponse
	endpoint := "/v2/payouts/fail-enqueued/" + url.PathEscape(payoutID)
	if err := c.doRequest(ctx, creds, http.MethodPost, endpoint, nil, &resp, "payout fail enqueued"); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) InitiateRefund(
	ctx context.Context,
	creds *Credentials,
	req *RefundRequest,
) (*RefundInitiationResponse, error) {
	payload := refundInitiationPayload{
		RefundID:          req.RefundID,
		DepositID:         req.DepositID,
		Amount:            req.Amount,
		Currency:          req.Currency,
		ClientReferenceID: req.ClientReferenceID,
		Metadata:          BuildMetadata(req.Metadata),
	}

	var resp RefundInitiationResponse
	if err := c.doRequest(ctx, creds, http.MethodPost, "/v2/refunds", payload, &resp, "refund"); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) GetRefund(
	ctx context.Context,
	creds *Credentials,
	refundID string,
) (*RefundStatusResult, error) {
	var resp RefundStatusResult
	endpoint := "/v2/refunds/" + url.PathEscape(refundID)
	if err := c.doRequest(ctx, creds, http.MethodGet, endpoint, nil, &resp, "refund status"); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) ResendRefundCallback(
	ctx context.Context,
	creds *Credentials,
	refundID string,
) (*ManualActionResponse, error) {
	var resp ManualActionResponse
	endpoint := "/v2/refunds/resend-callback/" + url.PathEscape(refundID)
	if err := c.doRequest(ctx, creds, http.MethodPost, endpoint, nil, &resp, "refund resend callback"); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) CancelEnqueuedRefund(
	ctx context.Context,
	creds *Credentials,
	refundID string,
) (*ManualActionResponse, error) {
	var resp ManualActionResponse
	endpoint := "/v2/refunds/fail-enqueued/" + url.PathEscape(refundID)
	if err := c.doRequest(ctx, creds, http.MethodPost, endpoint, nil, &resp, "refund fail enqueued"); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) CreatePaymentPageSession(
	ctx context.Context,
	creds *Credentials,
	req *PaymentPageRequest,
) (*PaymentPageSession, error) {
	payload := paymentPagePayload{
		DepositID:       req.DepositID,
		ReturnURL:       req.ReturnURL,
		CustomerMessage: req.CustomerMessage,
		PhoneNumber:     req.PhoneNumber,
		Language:        req.Language,
		Country:         req.Country,
		Reason:          req.Reason,
		Metadata:        BuildMetadata(req.Metadata),
	}
	if req.Amount != "" {
		payload.AmountDetails = &amountDetailsPayload{
			Amount:   req.Amount,
			Currency: req.Currency,
		}
	}

	var resp PaymentPageSession
	if err := c.doRequest(ctx, creds, http.MethodPost, "/v2/paymentpage", payload, &resp, "payment page"); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) PredictProvider(
	ctx context.Context,
	creds *Credentials,
	phoneNumber string,
) (*ProviderPrediction, error) {
	payload := map[string]string{"phoneNumber": phoneNumber}

	var resp ProviderPrediction
	if err := c.doRequest(
		ctx,
		creds,
		http.MethodPost,
		"/v2/predict-provider",
		payload,
		&resp,
		"predict provider",
	); err != nil {
		return nil, err
	}
	return &resp, nil
}

// filterQuery builds an optional ?country=&operationType= query string.
func filterQuery(country, operationType string) string {
	q := url.Values{}
	if country != "" {
		q.Set("country", country)
	}
	if operationType != "" {
		q.Set("operationType", operationType)
	}
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}

func (c *client) ActiveConfiguration(
	ctx context.Context,
	creds *Credentials,
	country, operationType string,
) (*ActiveConfiguration, error) {
	var resp ActiveConfiguration
	endpoint := "/v2/active-conf" + filterQuery(country, operationType)
	if err := c.doRequest(ctx, creds, http.MethodGet, endpoint, nil, &resp, "active configuration"); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) Availability(
	ctx context.Context,
	creds *Credentials,
	country, operationType string,
) ([]CountryAvailability, error) {
	var resp []CountryAvailability
	endpoint := "/v2/availability" + filterQuery(country, operationType)
	if err := c.doRequest(ctx, creds, http.MethodGet, endpoint, nil, &resp, "availability"); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *client) WalletBalances(
	ctx context.Context,
	creds *Credentials,
	country string,
) (*WalletBalances, error) {
	var resp WalletBalances
	endpoint := "/v2/wallet-balances" + filterQuery(country, "")
	if err := c.doRequest(ctx, creds, http.MethodGet, endpoint, nil, &resp, "wallet balances"); err != nil {
		return nil, err
	}
	return &resp, nil
}
