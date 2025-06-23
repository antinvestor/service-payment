package models

import (
	"time"

	"github.com/pitabwire/frame"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

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
	ID string `json:"id"`
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

// FetchBillersRequest represents the request structure for fetching billers
// It can be extended with additional fields if needed

type FetchBillersRequest struct {
	// Add fields if necessary
}

// TillsPayRequest represents the request structure for the tills/pay endpoint
type TillsPayRequest struct {
	Merchant TillsPayMerchant `json:"merchant"`
	Payment  TillsPayPayment  `json:"payment"`
	Partner  TillsPayPartner  `json:"partner"`
}

// TillsPayResponse represents the response structure for the tills/pay endpoint
type TillsPayResponse struct {
	Status    bool        `json:"status"`
	Code      int         `json:"code"`
	MerchantName string      `json:"merchantName"`
	TransactionID string    `json:"transactionId"`
	Message   string      `json:"message"`
}

type TillsPayMerchant struct {
	Till string `json:"till"`
}

type TillsPayPayment struct {
	Ref     string `json:"ref"`
	Amount  string `json:"amount"`
	Currency string `json:"currency"`
}

type TillsPayPartner struct {
	ID  string `json:"id"`
	Ref string `json:"ref"`
}

type Prompt struct {
	frame.BaseModel
	ID                string `gorm:"type:varchar(50)"`	
	SourceID          string `gorm:"type:varchar(50)"`
	SourceProfileType string `gorm:"type:varchar(50)"`
	SourceContactID   string `gorm:"type:varchar(50)"`

	RecipientID          string `gorm:"type:varchar(50)"`
	RecipientProfileType string `gorm:"type:varchar(50)"`
	RecipientContactID   string `gorm:"type:varchar(50)"`
	Amount           decimal.NullDecimal `gorm:"type:numeric" json:"amount"`
	DateCreated          string            `gorm:"type:varchar(50)"`
	DeviceID             string            `gorm:"type:varchar(50)"`
	State                int32             `gorm:"type:integer"`
	Status               int32             `gorm:"type:integer"`
	Route                string            `gorm:"type:varchar(50)"`
	Account     datatypes.JSON `gorm:"type:jsonb"`
	Extra       datatypes.JSONMap `gorm:"index:,type:gin;option:jsonb_path_ops" json:"extra"`
}

type Account struct {
	frame.BaseModel
	AccountNumber string `gorm:"type:varchar(50)"`
	CountryCode   string `gorm:"type:varchar(50)"`
	Name          string `gorm:"type:varchar(50)"`
}

type PaymentLink struct {
	frame.BaseModel

	ExpiryDate      time.Time       `gorm:"type:date" json:"expiryDate"`
	SaleDate        time.Time       `gorm:"type:date" json:"saleDate"`
	PaymentLinkType string          `gorm:"type:varchar(20)" json:"paymentLinkType"`
	SaleType        string          `gorm:"type:varchar(20)" json:"saleType"`
	Name            string          `gorm:"type:varchar(100)" json:"name"`
	Description     string          `gorm:"type:text" json:"description"`
	ExternalRef     string          `gorm:"type:varchar(50)" json:"externalRef"`
	PaymentLinkRef  string          `gorm:"type:varchar(50)" json:"paymentLinkRef"`
	RedirectURL     string          `gorm:"type:varchar(255)" json:"redirectURL"`
	AmountOption    string          `gorm:"type:varchar(20)" json:"amountOption"`
	Amount          decimal.Decimal `gorm:"type:numeric" json:"amount"`
	Currency        string          `gorm:"type:varchar(10)" json:"currency"`
	Customers       datatypes.JSON  `gorm:"type:jsonb" json:"customers"`     // stores []Customer as JSON
	Notifications   datatypes.JSON  `gorm:"type:jsonb" json:"notifications"` //Notifications are enums

}

// Payment Link API Models

type PaymentLinkCustomer struct {
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName"`
	Email             string `json:"email"`
	PhoneNumber       string `json:"phoneNumber"`
	FirstAddress      string `json:"firstAddress,omitempty"`
	CountryCode       string `json:"countryCode"`
	PostalOrZipCode   string `json:"postalOrZipCode,omitempty"`
	CustomerExternalRef string `json:"customerExternalRef,omitempty"`
}

type PaymentLinkDetails struct {
	ExpiryDate      string  `json:"expiryDate"`
	SaleDate        string  `json:"saleDate"`
	PaymentLinkType string  `json:"paymentLinkType"`
	SaleType        string  `json:"saleType"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	ExternalRef     string  `json:"externalRef"`
	PaymentLinkRef  string  `json:"paymentLinkRef,omitempty"`
	RedirectURL     string  `json:"redirectURL,omitempty"`
	AmountOption    string  `json:"amountOption"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
}

type PaymentLinkRequest struct {
	Customers     []PaymentLinkCustomer `json:"customers"`
	PaymentLink   PaymentLinkDetails    `json:"paymentLink"`
	Notifications []string              `json:"notifications,omitempty"`
}

type PaymentLinkResponse struct {
	Status   bool   `json:"status"`
	Code     int    `json:"code"`
	Message  string `json:"message"`
	Metadata map[string]interface{} `json:"metadata"`
	Data     *PaymentLinkResponseData `json:"data,omitempty"`
}

type PaymentLinkResponseData struct {
	DateCreated     int64  `json:"dateCreated"`
	PaymentLinkRef  string `json:"paymentLinkRef"`
	ExternalRef     string `json:"externalRef"`
	Status          struct {
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"status"`
}
