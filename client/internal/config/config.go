// Package config initialize configuration
package config

import (
	"github.com/caarlos0/env/v6"
)

// Config represents the configuration of the environment variable of the service
type Config struct {
	ServerHost string `env:"SERVER_HOST,notEmpty" envDefault:"localhost"`
	ServerPort string `env:"SERVER_PORT,notEmpty" envDefault:"12345"`
}

// New creates a new config of the service
func New() (*Config, error) {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, err
}
