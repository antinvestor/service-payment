package config

import "github.com/pitabwire/frame"

type JengaConfig struct {
	frame.ConfigurationDefault
	//JENGA_PRIVATE_KEY_PATH=/app/keys/privatekey.pem
	//JENGA_PUBLIC_KEY_PATH=/app/keys/publickey.pem
	JengaPrivateKey string `envDefault:"/app/keys/privatekey.pem" env:"JENGA_PRIVATE_KEY_PATH" required:"true"`
	ApiKey          string `envDefault:"SZq0WmmtX6mfo3fARW7yHeEzhfs3sOiEj2TgS2jb9gFz80JPfvTF1g4nr1uziA1meg3uFB1/Cm+ZXdTDob4z0Q==" env:"JENGA_API_KEY" required:"true"`
	ConsumerSecret  string `envDefault:"JZkt2pAIiS4F4RP4x6zQ97f1dn9j1N" env:"JENGA_CONSUMER_SECRET" required:"true"`
	MerchantCode    string `envDefault:"8503993262" env:"JENGA_MERCHANT_CODE" required:"true"`
	Env             string `envDefault:"https://uat.finserve.africa" env:"JENGA_ENV"`
	ProfileServiceURI   string `envDefault:"127.0.0.1:7005" env:"PROFILE_SERVICE_URI"`
	PartitionServiceURI string `envDefault:"127.0.0.1:7003" env:"PARTITION_SERVICE_URI"`
	SecurelyRunService  bool   `envDefault:"false" env:"SECURELY_RUN_SERVICE"`
	PaymentServiceURI    string `envDefault:"localhost:50051" env:"PAYMENT_SERVICE_URI" required:"true"`
	//NATS_URL=nats://${NATS_USER}:${NATS_PASSWORD}@nats-server:4222
	NATS_URL string `envDefault:"nats://${NATS_USER}:${NATS_PASSWORD}@nats-server:4222" env:"NATS_URL" required:"true"`
}
