package client

// PolarCredentials holds per-request credentials for Polar API.
type PolarCredentials struct {
	APIKey         string
	WebhookSecret  string
	OrganizationID string
	Environment    string
}

// BaseURL returns the Polar API base URL based on environment.
func (c *PolarCredentials) BaseURL() string {
	if c.Environment == "production" {
		return "https://api.polar.sh"
	}
	return "https://sandbox-api.polar.sh"
}

// CheckoutRequest represents a request to create a Polar checkout session.
type CheckoutRequest struct {
	ProductID     string
	CustomerEmail string
	Amount        int64
	Currency      string
	SuccessURL    string
	Metadata      map[string]string
}

// checkoutPayload is the JSON payload sent to the Polar API.
type checkoutPayload struct {
	ProductID        string            `json:"product_id,omitempty"`
	CustomerEmail    string            `json:"customer_email,omitempty"`
	Amount           int64             `json:"amount,omitempty"`
	Currency         string            `json:"currency,omitempty"`
	SuccessURL       string            `json:"success_url,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	PaymentProcessor string            `json:"payment_processor"`
}

// CheckoutResponse represents the response from creating a checkout session.
type CheckoutResponse struct {
	ID           string `json:"id"`
	URL          string `json:"url"`
	ClientSecret string `json:"client_secret"`
	Status       string `json:"status"`
}

// WebhookEvent represents a parsed Polar webhook event.
type WebhookEvent struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}
