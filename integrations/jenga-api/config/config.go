package config

type JengaConfig struct {
	APIKey         string `default:"" envconfig:"JENGA_API_KEY"`
	ConsumerSecret string `default:"" envconfig:"JENGA_CONSUMER_SECRET"`
	MerchantCode   string `default:"" envconfig:"JENGA_MERCHANT_CODE"`
	Env            string `default:"" envconfig:"JENGA_ENV"`
}
