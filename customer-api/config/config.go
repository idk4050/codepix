package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/caarlos0/env"
	"github.com/subosito/gotenv"
)

type Config struct {
	Database database
	HTTP     http
}

func New() (*Config, error) {
	c := &Config{
		Database: database{},
		HTTP:     http{},
	}
	err := loadEnvFileIfAvailable()
	if err != nil {
		return nil, err
	}
	env.Parse(&c.Database)
	if c.Database == (database{}) {
		return nil, errors.New("failed to load database config")
	}
	env.Parse(&c.HTTP)
	if c.HTTP == (http{}) {
		return nil, errors.New("failed to load HTTP config")
	}
	return c, nil
}

func loadEnvFileIfAvailable() error {
	_, thisFile, _, _ := runtime.Caller(0)
	envFilePath := filepath.Join(filepath.Dir(thisFile), "./env/.env")

	if _, err := os.Stat(envFilePath); err == nil {
		err := gotenv.Load(envFilePath)
		if err != nil {
			return fmt.Errorf("load env files: %w", err)
		}
	}
	return nil
}

type database struct {
	Dialect          string `env:"DB_DIALECT"`
	ConnectionString string `env:"DB_CONNECTION_STRING"`
	Host             string `env:"DB_HOST"`
	Port             string `env:"DB_PORT"`
	Name             string `env:"DB_NAME"`
	User             string `env:"DB_USER"`
	Password         string `env:"DB_PASSWORD"`
	SSLMode          string `env:"DB_SSLMODE"`
	AutoMigrate      bool   `env:"DB_AUTO_MIGRATE"`
}

type http struct {
	Port string `env:"HTTP_PORT"`
}
