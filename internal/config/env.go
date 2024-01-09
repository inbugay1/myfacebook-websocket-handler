package config

import (
	"github.com/caarlos0/env/v6"
	"log"
)

type EnvConfig struct {
	Version  string `env:"VERSION" envDefault:"version_not_set"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
	HTTPPort string `env:"HTTP_INT_PORT" envDefault:"9093"`

	RequestHeaderMaxSize                 int `env:"REQUEST_HEADER_MAX_SIZE" envDefault:"10000"`
	RequestReadHeaderTimeoutMilliseconds int `env:"REQUEST_READ_HEADER_TIMEOUT_MILLISECONDS" envDefault:"2000"`
}

func GetConfigFromEnv() *EnvConfig {
	var config EnvConfig

	if err := env.Parse(&config); err != nil {
		log.Fatalf("unable to parse env config, error: %s", err)
	}

	return &config
}
