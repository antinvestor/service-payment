package client

// MtnCredentials holds the per-request credentials for MTN MoMo API calls.
type MtnCredentials struct {
	SubscriptionKey string
	APIUser         string
	APIKey          string
	CallbackURL     string
	Currency        string
	Environment     string // sandbox or production
}

// BaseURL returns the MTN MoMo API base URL based on environment.
func (c *MtnCredentials) BaseURL() string {
	if c.Environment == "production" {
		return "https://proxy.momoapi.mtn.com"
	}
	return "https://sandbox.momodeveloper.mtn.com"
}

// RequestToPayRequest represents a request-to-pay (collection) request.
type RequestToPayRequest struct {
	ReferenceID  string
	Amount       string
	Currency     string
	ExternalID   string
	Payer        Party
	PayerMessage string
	PayeeNote    string
	CallbackURL  string
}

// requestToPayPayload is the JSON payload sent to MTN MoMo API.
type requestToPayPayload struct {
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	ExternalID   string `json:"externalId"`
	Payer        Party  `json:"payer"`
	PayerMessage string `json:"payerMessage"`
	PayeeNote    string `json:"payeeNote"`
}

// Party represents a party in a transaction.
type Party struct {
	PartyIDType string `json:"partyIdType"`
	PartyID     string `json:"partyId"`
}

// RequestToPayStatus represents the status of a request-to-pay.
type RequestToPayStatus struct {
	Amount                 string `json:"amount"`
	Currency               string `json:"currency"`
	FinancialTransactionID string `json:"financialTransactionId"`
	ExternalID             string `json:"externalId"`
	Payer                  Party  `json:"payer"`
	Status                 string `json:"status"` // PENDING, SUCCESSFUL, FAILED
	Reason                 *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"reason"`
}

// TransferRequest represents a disbursement transfer request.
type TransferRequest struct {
	ReferenceID  string
	Amount       string
	Currency     string
	ExternalID   string
	Payee        Party
	PayerMessage string
	PayeeNote    string
	CallbackURL  string
}

// transferPayload is the JSON payload for a disbursement transfer.
type transferPayload struct {
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	ExternalID   string `json:"externalId"`
	Payee        Party  `json:"payee"`
	PayerMessage string `json:"payerMessage"`
	PayeeNote    string `json:"payeeNote"`
}

// TransferStatus represents the status of a disbursement transfer.
type TransferStatus struct {
	Amount                 string `json:"amount"`
	Currency               string `json:"currency"`
	FinancialTransactionID string `json:"financialTransactionId"`
	ExternalID             string `json:"externalId"`
	Payee                  Party  `json:"payee"`
	Status                 string `json:"status"` // PENDING, SUCCESSFUL, FAILED
	Reason                 *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"reason"`
}

// CallbackBody represents a MTN MoMo callback notification.
type CallbackBody struct {
	FinancialTransactionID string `json:"financialTransactionId"`
	ExternalID             string `json:"externalId"`
	Amount                 string `json:"amount"`
	Currency               string `json:"currency"`
	Payer                  *Party `json:"payer"`
	Payee                  *Party `json:"payee"`
	Status                 string `json:"status"`
	Reason                 *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"reason"`
}
