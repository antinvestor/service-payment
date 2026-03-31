package client

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/pitabwire/util"
)

const (
	httpClientTimeout         = 30 * time.Second
	webhookTimestampTolerance = 300 // seconds
	signatureParts            = 2
)

type polarClient struct {
	httpClient *http.Client
}

// NewClient creates a new Polar.sh API client.
func NewClient() PolarClient {
	return &polarClient{
		httpClient: &http.Client{
			Timeout: httpClientTimeout,
		},
	}
}

func (c *polarClient) CreateCheckout(
	ctx context.Context,
	creds *PolarCredentials,
	req *CheckoutRequest,
) (*CheckoutResponse, error) {
	logger := util.Log(ctx).WithField("type", "polar.create_checkout")
	defer logger.Release()

	payload := checkoutPayload{
		ProductID:        req.ProductID,
		CustomerEmail:    req.CustomerEmail,
		Amount:           req.Amount,
		Currency:         req.Currency,
		SuccessURL:       req.SuccessURL,
		Metadata:         req.Metadata,
		PaymentProcessor: "stripe",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal checkout payload: %w", err)
	}

	url := creds.BaseURL() + "/v1/checkouts/"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create checkout request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+creds.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute checkout request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	logger.WithFields(map[string]any{"status": resp.StatusCode, "response": string(respBody)}).Debug("checkout response")

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("checkout creation failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var checkoutResp CheckoutResponse
	if err = json.Unmarshal(respBody, &checkoutResp); err != nil {
		return nil, fmt.Errorf("decode checkout response: %w", err)
	}

	return &checkoutResp, nil
}

func (c *polarClient) VerifyWebhookSignature(
	payload []byte,
	headers map[string]string,
	webhookSecret string,
) (*WebhookEvent, error) {
	// Polar uses Standard Webhooks (svix) format
	webhookID := headers["webhook-id"]
	if webhookID == "" {
		webhookID = headers["Webhook-Id"]
	}

	timestamp := headers["webhook-timestamp"]
	if timestamp == "" {
		timestamp = headers["Webhook-Timestamp"]
	}

	signature := headers["webhook-signature"]
	if signature == "" {
		signature = headers["Webhook-Signature"]
	}

	if webhookID == "" || timestamp == "" || signature == "" {
		return nil, errors.New("missing required webhook headers (webhook-id, webhook-timestamp, webhook-signature)")
	}

	// Verify timestamp to prevent replay attacks (5 minute tolerance)
	ts, err := parseTimestamp(timestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook timestamp: %w", err)
	}
	if math.Abs(time.Since(ts).Seconds()) > webhookTimestampTolerance {
		return nil, errors.New("webhook timestamp too old or too new")
	}

	// Standard Webhooks signature: base64(hmac-sha256(secret, "{msg_id}.{timestamp}.{body}"))
	// The secret is base64-encoded with "whsec_" prefix
	secretBytes, err := decodeWebhookSecret(webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("decode webhook secret: %w", err)
	}

	signedContent := fmt.Sprintf("%s.%s.%s", webhookID, timestamp, string(payload))
	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(signedContent))
	expectedSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Signature header may contain multiple signatures separated by spaces
	// Each signature has format "v1,{base64sig}"
	verified := false
	for _, sig := range strings.Split(signature, " ") {
		parts := strings.SplitN(sig, ",", signatureParts)
		if len(parts) == 2 && parts[1] == expectedSig {
			verified = true
			break
		}
	}

	if !verified {
		return nil, errors.New("webhook signature verification failed")
	}

	var event WebhookEvent
	if unmarshalErr := json.Unmarshal(payload, &event); unmarshalErr != nil {
		return nil, fmt.Errorf("decode webhook event: %w", unmarshalErr)
	}

	return &event, nil
}

func parseTimestamp(ts string) (time.Time, error) {
	// Standard Webhooks timestamp is Unix epoch seconds
	var epoch int64
	if _, err := fmt.Sscanf(ts, "%d", &epoch); err != nil {
		return time.Time{}, err
	}
	return time.Unix(epoch, 0), nil
}

func decodeWebhookSecret(secret string) ([]byte, error) {
	// Standard Webhooks secrets have "whsec_" prefix followed by base64-encoded key
	secret = strings.TrimPrefix(secret, "whsec_")
	return base64.StdEncoding.DecodeString(secret)
}
