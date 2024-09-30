package config

import "github.com/pitabwire/frame"

type PaymentConfig struct {
	frame.ConfigurationDefault
	ProfileServiceURI   string `default:"127.0.0.1:7005" envconfig:"PROFILE_SERVICE_URI"`
	PartitionServiceURI string `default:"127.0.0.1:7003" envconfig:"PARTITION_SERVICE_URI"`

	SecurelyRunService bool `default:"true" envconfig:"SECURELY_RUN_SERVICE"`
}
