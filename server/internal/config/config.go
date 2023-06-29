// Package config initialize configuration
package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

// Environment - run application environment
type Environment string

const (
	// Develop - development environment
	Develop Environment = "DEV"
	// Production - production environment
	Production Environment = "PROD"
)

// Config represents the configuration of the environment variable of the service
type Config struct {
	ServiceName string      `env:"SERVICE_NAME,notEmpty" envDefault:"Word of Wisdom"`
	ServiceHost string      `env:"SERVICE_HOST,notEmpty" envDefault:"0.0.0.0"`
	ServicePort string      `env:"SERVICE_PORT,notEmpty" envDefault:"12345"`
	Environment Environment `env:"ENVIRONMENT,notEmpty" envDefault:"PROD"`

	Sqlite

	Pow
}

// New creates a new config of the service
func New() (*Config, error) {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	err = cfg.validate()
	if err != nil {
		return nil, err
	}

	return cfg, err
}

func (c *Config) validate() error {
	switch c.Sqlite.SQLiteMode {
	case SqliteOpenCreate, SqliteOpenReadonly, SqliteOpenReadwrite, SqliteOpenMemory:
	default:
		return fmt.Errorf(`specified SQLiteMode doesn't exist`)
	}

	return nil
}
