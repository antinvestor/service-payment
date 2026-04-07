package models

import "encoding/json"

// StkCallback represents the STK push callback from Jenga.
// Amount fields use json.Number to preserve exact decimal representation and avoid float64 precision loss.
type StkCallback struct {
	Status        bool        `json:"status"`
	Code          int         `json:"code"`
	Message       string      `json:"message"`
	Transaction   string      `json:"transactionReference"`
	Telco         string      `json:"telcoReference"`
	MobileNumber  string      `json:"mobileNumber"`
	Currency      string      `json:"currency"`
	RequestAmount json.Number `json:"requestAmount"`
	DebitedAmount json.Number `json:"debitedAmount"`
	Charge        json.Number `json:"charge"`
	TelcoName     string      `json:"telco"`
}

// CallbackRequest represents a general payment callback from Jenga.
type CallbackRequest struct {
	CallbackType string `json:"callbackType"`
	Customer     struct {
		Name         string `json:"name"`
		MobileNumber string `json:"mobileNumber"`
		Reference    string `json:"reference"`
	} `json:"customer"`
	Transaction struct {
		Date           string      `json:"date"`
		Reference      string      `json:"reference"`
		PaymentMode    string      `json:"paymentMode"`
		Amount         json.Number `json:"amount"`
		Currency       string      `json:"currency"`
		BillNumber     string      `json:"billNumber"`
		ServedBy       string      `json:"servedBy"`
		AdditionalInfo string      `json:"additionalInfo"`
		OrderAmount    json.Number `json:"orderAmount"`
		ServiceCharge  json.Number `json:"serviceCharge"`
		OrderCurrency  string      `json:"orderCurrency"`
		Status         string      `json:"status"`
		Remarks        string      `json:"remarks"`
	} `json:"transaction"`
	Bank struct {
		Reference       string `json:"reference"`
		TransactionType string `json:"transactionType"`
		Account         string `json:"account"`
	} `json:"bank"`
}
