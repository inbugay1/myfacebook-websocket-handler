package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type EnvConfig struct {
	Version  string `env:"VERSION" envDefault:"version_not_set"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
	HTTPPort string `env:"HTTP_INT_PORT" envDefault:"9093"`

	RequestHeaderMaxSize                 int `env:"REQUEST_HEADER_MAX_SIZE" envDefault:"10000"`
	RequestReadHeaderTimeoutMilliseconds int `env:"REQUEST_READ_HEADER_TIMEOUT_MILLISECONDS" envDefault:"2000"`

	MyfacebookAPIBaseURL string `env:"MYFACEBOOK_API_BASE_URL" envDefault:"http://localhost:9092"`

	WebSocketUpgraderReadBufferSize  int  `env:"WEB_SOCKET_UPGRADER_READ_BUFFER_SIZE" envDefault:"2048"`
	WebSocketUpgraderWriteBufferSize int  `env:"WEB_SOCKET_UPGRADER_WRITE_BUFFER_SIZE" envDefault:"2048"`
	WebSocketUpgraderCheckOrigin     bool `env:"WEB_SOCKET_UPGRADER_CHECK_ORIGIN" envDefault:"true"`

	WebSocketMaxMessageSize      int `env:"WEB_SOCKET_MAX_MESSAGE_SIZE" envDefault:"2048"`
	WebSocketReadTimeoutSeconds  int `env:"WEB_SOCKET_READ_TIMEOUT_SECONDS" envDefault:"60"`
	WebSocketWriteTimeoutSeconds int `env:"WEB_SOCKET_WRITE_TIMEOUT_SECONDS" envDefault:"5"`
	WebSocketPingIntervalSeconds int `env:"WEB_SOCKET_PING_INTERVAL_SECONDS" envDefault:"10"`

	RMQHost     string `env:"RMQ_HOST" envDefault:"localhost"`
	RMQPort     string `env:"RMQ_PORT" envDefault:"5672"`
	RMQUsername string `env:"RMQ_USERNAME" envDefault:"guest"`
	RMQPassword string `env:"RMQ_PASSWORD" envDefault:"guest"`

	ConnectionWatcherPingIntervalSeconds     int `env:"CONNECTION_WATCHER_PING_INTERVAL_SECONDS" envDefault:"5"`
	ConnectionWatcherPingTimeoutSeconds      int `env:"CONNECTION_WATCHER_PING_TIMEOUT_SECONDS" envDefault:"2"`
	ConnectionWatcherReconnectTimeoutSeconds int `env:"CONNECTION_WATCHER_RECONNECT_TIMEOUT_SECONDS" envDefault:"2"`
}

func GetConfigFromEnv() *EnvConfig {
	var config EnvConfig

	if err := env.Parse(&config); err != nil {
		log.Fatalf("unable to parse env config, error: %s", err)
	}

	return &config
}
