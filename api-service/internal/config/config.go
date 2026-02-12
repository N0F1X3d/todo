package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServiceName string `env:"SERVICE_NAME" env-default:"todo-api-service"`
	Environment string `env:"ENVIRONMENT" env-default:"development"`

	// HTTP Server
	HTTPHost         string        `env:"HTTP_HOST" env-default:"0.0.0.0"`
	HTTPPort         int           `env:"HTTP_PORT" env-default:"8080"`
	HTTPReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"10s"`
	HTTPWriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`

	// gRPC Client (db-service)
	GRPCHost string `env:"GRPC_HOST" env-default:"localhost"`
	GRPCPort int    `env:"GRPC_PORT" env-default:"50051"`

	// Kafka (будущее)
	KafkaBrokers string `env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	KafkaTopic   string `env:"KAFKA_TOPIC" env-default:"todo-events"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) HTTPAddress() string {
	return fmt.Sprintf("%s:%d", c.HTTPHost, c.HTTPPort)
}

func (c *Config) GRPCAddress() string {
	return fmt.Sprintf("%s:%d", c.GRPCHost, c.GRPCPort)
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
