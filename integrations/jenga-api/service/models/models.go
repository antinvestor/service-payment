package models
type Merchant struct {
	CountryCode   string `json:"countryCode"`
	AccountNumber string `json:"accountNumber"`
	Name          string `json:"name"`
}

type Payment struct {
	Ref         string `json:"ref"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Telco       string `json:"telco"`
	MobileNumber string `json:"mobileNumber"`
	Date        string `json:"date"`
	CallBackUrl string `json:"callBackUrl"`
	PushType    string `json:"pushType"`
}
