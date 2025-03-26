package config
import "github.com/pitabwire/frame"

type JengaConfig struct {
	//JENGA_PRIVATE_KEY_PATH=/app/keys/privatekey.pem
	//JENGA_PUBLIC_KEY_PATH=/app/keys/publickey.pem

	JengaPrivateKey string `default:"/app/keys/privatekey.pem" envconfig:"JENGA_PRIVATE_KEY_PATH"`
	ApiKey         string `default:"SZq0WmmtX6mfo3fARW7yHeEzhfs3sOiEj2TgS2jb9gFz80JPfvTF1g4nr1uziA1meg3uFB1/Cm+ZXdTDob4z0Q==" envconfig:"JENGA_API_KEY"`
	ConsumerSecret string `default:"JZkt2pAIiS4F4RP4x6zQ97f1dn9j1N" envconfig:"JENGA_CONSUMER_SECRET"`
	MerchantCode   string `default:"8503993262" envconfig:"JENGA_MERCHANT_CODE"`
	Env            string `default:"https://uat.finserve.africa" envconfig:"JENGA_ENV"`
	frame.ConfigurationDefault
	ProfileServiceURI   string `default:"127.0.0.1:7005" envconfig:"PROFILE_SERVICE_URI"`
	PartitionServiceURI string `default:"127.0.0.1:7003" envconfig:"PARTITION_SERVICE_URI"`
	SecurelyRunService bool `default:"false" envconfig:"SECURELY_RUN_SERVICE"`
	//redis
	RedisHost string `default:"localhost" envconfig:"REDIS_HOST"`
	RedisPort string `default:"6379" envconfig:"REDIS_PORT"`

	
}
