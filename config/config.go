package config

import "github.com/pitabwire/frame"

type PaymentConfig struct {
	frame.ConfigurationDefault
	ProfileServiceURI   string `envDefault:"127.0.0.1:7005" env:"PROFILE_SERVICE_URI"`
	PartitionServiceURI string `envDefault:"127.0.0.1:7003" env:"PARTITION_SERVICE_URI"`

	SecurelyRunService bool `envDefault:"true" env:"SECURELY_RUN_SERVICE"`
}
