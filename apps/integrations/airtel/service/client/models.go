package client

// AirtelCredentials holds per-request credentials for Airtel Money API.
type AirtelCredentials struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
	CountryCode  string
	Currency     string
	Environment  string
}

// BaseURL returns the Airtel Money API base URL based on environment.
func (c *AirtelCredentials) BaseURL() string {
	if c.Environment == "production" {
		return "https://openapi.airtel.africa"
	}
	return "https://openapiuat.airtel.africa"
}

// TokenResponse represents the OAuth token response.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// CollectionRequest represents a USSD push collection request.
type CollectionRequest struct {
	Reference   string
	PhoneNumber string
	Amount      string
	Currency    string
	CountryCode string
	CallbackURL string
}

// collectionPayload is the JSON payload sent to the Airtel API.
type collectionPayload struct {
	Reference   string             `json:"reference"`
	Subscriber  subscriberPayload  `json:"subscriber"`
	Transaction transactionPayload `json:"transaction"`
}

type subscriberPayload struct {
	Country  string `json:"country"`
	Currency string `json:"currency"`
	Msisdn   string `json:"msisdn"`
}

type transactionPayload struct {
	Amount   string `json:"amount"`
	Country  string `json:"country"`
	Currency string `json:"currency"`
	ID       string `json:"id"`
}

// CollectionResponse represents the API response for a collection request.
type CollectionResponse struct {
	Data struct {
		Transaction struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"transaction"`
	} `json:"data"`
	Status struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		ResultCode string `json:"result_code"`
		Success    bool   `json:"success"`
	} `json:"status"`
}

// DisbursementRequest represents a disbursement request.
type DisbursementRequest struct {
	Reference   string
	PhoneNumber string
	Amount      string
	Currency    string
	CountryCode string
}

// disbursementPayload is the JSON payload sent to the Airtel API.
type disbursementPayload struct {
	Payee       payeePayload       `json:"payee"`
	Reference   string             `json:"reference"`
	Pin         string             `json:"pin"`
	Transaction transactionPayload `json:"transaction"`
}

type payeePayload struct {
	Msisdn string `json:"msisdn"`
	Name   string `json:"name,omitempty"`
}

// DisbursementResponse represents the API response for a disbursement.
type DisbursementResponse struct {
	Data struct {
		Transaction struct {
			ID          string `json:"id"`
			Status      string `json:"status"`
			ReferenceID string `json:"reference_id"`
		} `json:"transaction"`
	} `json:"data"`
	Status struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		ResultCode string `json:"result_code"`
		Success    bool   `json:"success"`
	} `json:"status"`
}

// StatusResponse represents the transaction status API response.
type StatusResponse struct {
	Data struct {
		Transaction struct {
			AirtelMoneyID string `json:"airtel_money_id"`
			ID            string `json:"id"`
			Message       string `json:"message"`
			Status        string `json:"status"`
		} `json:"transaction"`
	} `json:"data"`
	Status struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		ResultCode string `json:"result_code"`
		Success    bool   `json:"success"`
	} `json:"status"`
}

// CollectionCallbackBody represents the callback payload from Airtel for collections.
type CollectionCallbackBody struct {
	Transaction struct {
		ID          string `json:"id"`
		Message     string `json:"message"`
		StatusCode  string `json:"status_code"`
		AirtelMoney struct {
			ID string `json:"id"`
		} `json:"airtel_money_id"`
	} `json:"transaction"`
}

// DisbursementCallbackBody represents the callback payload from Airtel for disbursements.
type DisbursementCallbackBody struct {
	Transaction struct {
		ID          string `json:"id"`
		Message     string `json:"message"`
		StatusCode  string `json:"status_code"`
		AirtelMoney struct {
			ID string `json:"id"`
		} `json:"airtel_money_id"`
	} `json:"transaction"`
}
