package models

type CallbackRequest struct {
	CallbackType string `json:"callbackType"`
	Customer     struct {
		Name         string `json:"name"`
		MobileNumber string `json:"mobileNumber"`
		Reference    string `json:"reference"`
	} `json:"customer"`
	Transaction struct {
		Date           string  `json:"date"`
		Reference      string  `json:"reference"`
		PaymentMode    string  `json:"paymentMode"`
		Amount         float64 `json:"amount"`
		Currency       string  `json:"currency"`
		BillNumber     string  `json:"billNumber"`
		ServedBy       string  `json:"servedBy"`
		AdditionalInfo string  `json:"additionalInfo"`
		OrderAmount    float64 `json:"orderAmount"`
		ServiceCharge  float64 `json:"serviceCharge"`
		OrderCurrency  string  `json:"orderCurrency"`
		Status         string  `json:"status"`
		Remarks        string  `json:"remarks"`
	} `json:"transaction"`
	Bank struct {
		Reference       string `json:"reference"`
		TransactionType string `json:"transactionType"`
		Account        *string `json:"account"`
	} `json:"bank"`
}
