package models

type Job struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	ExtraData interface{} `json:"extra_data"`
}

type Callback struct {
	CallbackUrl string `json:"callback_url"`
}
