package core_auth

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	JWTSecret string        `envconfig:"JWT_SECRET" required:"true"`
	TokenTTL  time.Duration `envconfig:"TOKEN_TTL" default:"24h"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("AUTH", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	if len(config.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("AUTH_JWT_SECRET must be 32bytes, got '%d'", len(config.JWTSecret))
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get auth config: %w", err)
		panic(err)
	}

	return config
}
