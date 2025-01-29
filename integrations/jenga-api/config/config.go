package config
import "github.com/pitabwire/frame"

type JengaConfig struct {
	ApiKey         string `default:"" envconfig:"JENGA_API_KEY"`
	ConsumerSecret string `default:"" envconfig:"JENGA_CONSUMER_SECRET"`
	MerchantCode   string `default:"" envconfig:"JENGA_MERCHANT_CODE"`
	Env            string `default:"" envconfig:"JENGA_ENV"`
	frame.ConfigurationDefault
	ProfileServiceURI   string `default:"127.0.0.1:7005" envconfig:"PROFILE_SERVICE_URI"`
	PartitionServiceURI string `default:"127.0.0.1:7003" envconfig:"PARTITION_SERVICE_URI"`
	SecurelyRunService bool `default:"false" envconfig:"SECURELY_RUN_SERVICE"`

	
}
