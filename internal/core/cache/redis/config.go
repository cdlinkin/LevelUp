package core_redis_cache

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr    string        `envconfig:"ADDR" required:"true"`
	Timeout time.Duration `envconfig:"TIMEOUT" required:"true"`
	TTL     time.Duration `envconfig:"TTL" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("REDIS", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	if config.Timeout <= 0 {
		return Config{}, fmt.Errorf("timeout must be positive")
	}

	if config.TTL <= 0 {
		return Config{}, fmt.Errorf("ttl must be positive")
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get redis config: %w", err)
		panic(err)
	}

	return config
}
