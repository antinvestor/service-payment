package models

type StkCallback struct {
	Status        bool    `json:"status"`
	Code          int     `json:"code"`
	Message       string  `json:"message"`
	Transaction   string  `json:"transactionReference"`
	Telco         string  `json:"telcoReference"`
	MobileNumber  string  `json:"mobileNumber"`
	Currency      string  `json:"currency"`
	RequestAmount float64 `json:"requestAmount"`
	DebitedAmount float64 `json:"debitedAmount"`
	Charge        float64 `json:"charge"`
	TelcoName     string  `json:"telco"`
}
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
		Account         string `json:"account"`
	} `json:"bank"`
}
