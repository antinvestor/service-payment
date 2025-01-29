package config
import "github.com/pitabwire/frame"

type JengaConfig struct {
	//JENGA_PRIVATE_KEY_PATH=/app/keys/privatekey.pem
	//JENGA_PUBLIC_KEY_PATH=/app/keys/publickey.pem

	JengaPrivateKey string `default:"/app/keys/privatekey.pem" envconfig:"JENGA_PRIVATE_KEY_PATH"`
	ApiKey         string `default:"r/PSfbyIYe/0epGhnawG1+g96mHA5zRtQaFc54ySQWAJIMyoLLDZSczc9h/gVrIHreueBDU2kRqH7clhMDEA9Q==" envconfig:"JENGA_API_KEY"`
	ConsumerSecret string `default:"iS3N5hDNwGK3Zm9F1eJ2Aic64e4da8" envconfig:"JENGA_CONSUMER_SECRET"`
	MerchantCode   string `default:"8503993262" envconfig:"JENGA_MERCHANT_CODE"`
	Env            string `default:"https://uat.finserve.africa" envconfig:"JENGA_ENV"`
	frame.ConfigurationDefault
	ProfileServiceURI   string `default:"127.0.0.1:7005" envconfig:"PROFILE_SERVICE_URI"`
	PartitionServiceURI string `default:"127.0.0.1:7003" envconfig:"PARTITION_SERVICE_URI"`
	SecurelyRunService bool `default:"false" envconfig:"SECURELY_RUN_SERVICE"`

	
}
