package models

type Merchant struct {
	CountryCode   string `json:"countryCode"`
	AccountNumber string `json:"accountNumber"`
	Name          string `json:"name"`
}

type PaymentRequest struct {
	Biller    Biller `json:"biller"`
	Bill      Bill   `json:"bill"`
	Payer     Payer  `json:"payer"`
	PartnerID string `json:"partnerId"`
	Remarks   string `json:"remarks"`
}

// PaymentResponse represents the response structure for the payment request

type PaymentResponse struct {
	Status    bool   `json:"status"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Reference string `json:"reference"`
	Data      struct {
		TransactionId string `json:"transactionId"`
		Status        string `json:"status"`
	} `json:"data"`
}

// STKUSSDRequest represents the structure for the STK/USSD push request
type STKUSSDRequest struct {
	Merchant Merchant `json:"merchant"`
	Payment  Payment  `json:"payment"`
}

// STKUSSDResponse represents the response structure for the STK/USSD push initiation
type STKUSSDResponse struct {
	Status        bool   `json:"status"`
	Code          int    `json:"code"`
	Message       string `json:"message"`
	Reference     string `json:"reference"`
	TransactionID string `json:"transactionId"`
}

type Payment struct {
	Ref          string `json:"ref"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	Telco        string `json:"telco"`
	MobileNumber string `json:"mobileNumber"`
	Date         string `json:"date"`
	CallBackUrl  string `json:"callBackUrl"`
	PushType     string `json:"pushType"`
}

type Biller struct {
	BillerCode  string `json:"billerCode"`
	CountryCode string `json:"countryCode"`
}

type Bill struct {
	Reference string `json:"reference"`
	Amount    string `json:"amount"`
	Currency  string `json:"currency"`
}

type Payer struct {
	Name         string `json:"name"`
	Account      string `json:"account"`
	Reference    string `json:"reference"`
	MobileNumber string `json:"mobileNumber"`
}

type Job struct {
	ID        string         `json:"id"`
	ExtraData PaymentRequest `json:"extra_data"`
}

type AccountBalanceRequest struct {
	CountryCode string `json:"countryCode"`
	AccountId     string `json:"account"`
}

//BalanceResponse represents the response structure for the account balance

type BalanceResponse struct {
	Status  bool   `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Balances []struct {
			Amount string `json:"amount"`
			Type   string `json:"type"`
		} `json:"balances"`
		Currency string `json:"currency"`
	} `json:"data"`
}
