package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"

	"github.com/caarlos0/env"
	"github.com/subosito/gotenv"
)

type Config struct {
	Database   database
	EventStore eventStore
}

func New() (*Config, error) {
	c := &Config{
		Database:   database{},
		EventStore: eventStore{},
	}
	err := loadEnvFileIfAvailable()
	if err != nil {
		return nil, err
	}
	env.Parse(&c.Database)
	if c.Database == (database{}) {
		return nil, errors.New("failed to load database config")
	}
	env.Parse(&c.EventStore)
	if reflect.DeepEqual(c.EventStore, eventStore{}) {
		return nil, errors.New("failed to load event store config")
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

type eventStore struct {
	InMemory       bool     `env:"ES_IN_MEMORY"`
	ReplicaSetName string   `env:"ES_REPLICA_SET_NAME"`
	Hosts          []string `env:"ES_HOSTS"`
	Name           string   `env:"ES_NAME"`
	User           string   `env:"ES_USER"`
	Password       string   `env:"ES_PASSWORD"`
}
