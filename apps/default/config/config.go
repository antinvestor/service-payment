package config

import "github.com/pitabwire/frame"

type PaymentConfig struct {
	frame.ConfigurationDefault
	ProfileServiceURI   string `envDefault:"127.0.0.1:7005" env:"PROFILE_SERVICE_URI"`
	PartitionServiceURI string `envDefault:"127.0.0.1:7003" env:"PARTITION_SERVICE_URI"`

	SecurelyRunService bool `envDefault:"true" env:"SECURELY_RUN_SERVICE"`
	NATS_URL string `envDefault:"nats://ant:secret@nats-server:4222?subject=" env:"NATS_URL" required:"true"`
	PromptTopic string `envDefault:"initiate.prompt" env:"PROMPT_TOPIC" required:"true"`
	PaymentLinkTopic string `envDefault:"create.payment.link" env:"PAYMENT_LINK_TOPIC" required:"true"`
	DO_MIGRATION bool `envDefault:"false" env:"DO_MIGRATION"`
	// The callback URL for Jenga STK push notifications
	//DATABASE_URL string `envDefault:"postgres://ant:secret@payment_db:5432/service_payment?sslmode=disable" env:"DATABASE_URL" required:"true"`

}
