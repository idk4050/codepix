package config

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/caarlos0/env"
	"github.com/subosito/gotenv"
)

type Config struct {
	Database        database
	EventStore      eventStore
	StoreProjection storeProjection
	EventBus        eventBus
	RPC             rpc
	BankAuth        bankAuth
}

func New() (*Config, error) {
	c := &Config{
		Database:        database{},
		EventStore:      eventStore{},
		StoreProjection: storeProjection{},
		EventBus:        eventBus{},
		RPC:             rpc{},
		BankAuth:        bankAuth{},
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
	env.Parse(&c.StoreProjection)
	if reflect.DeepEqual(c.StoreProjection, storeProjection{}) {
		return nil, errors.New("failed to load store projection config")
	}
	env.Parse(&c.EventBus)
	if c.EventBus == (eventBus{}) {
		return nil, errors.New("failed to load event bus config")
	}
	env.Parse(&c.RPC)
	if c.RPC == (rpc{}) {
		return nil, errors.New("failed to load RPC config")
	}
	env.Parse(&c.BankAuth)
	if c.BankAuth == (bankAuth{}) {
		return nil, errors.New("failed to load bank auth config")
	}
	err = c.BankAuth.build()
	if err != nil {
		return nil, fmt.Errorf("failed to build bank auth config: %w", err)
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

type storeProjection struct {
	InMemory       bool     `env:"SP_IN_MEMORY"`
	ReplicaSetName string   `env:"SP_REPLICA_SET_NAME"`
	Hosts          []string `env:"SP_HOSTS"`
	Name           string   `env:"SP_NAME"`
	User           string   `env:"SP_USER"`
	Password       string   `env:"SP_PASSWORD"`
}

type eventBus struct {
	InMemory bool   `env:"EB_IN_MEMORY"`
	Host     string `env:"EB_HOST"`
	Port     string `env:"EB_PORT"`
	Name     string `env:"EB_NAME"`
	User     string `env:"EB_USER"`
	Password string `env:"EB_PASSWORD"`
}

type rpc struct {
	Port string `env:"RPC_PORT"`
}

type bankAuth struct {
	ValidationKey               any
	ValidationKeyString         string `env:"BANK_AUTH_VALIDATION_KEY"`
	PreviousValidationKey       any
	PreviousValidationKeyString string `env:"BANK_AUTH_PREVIOUS_VALIDATION_KEY"`
}

func (c *bankAuth) build() error {
	validationKeyPem, _ := pem.Decode([]byte(escapeNewLines(c.ValidationKeyString)))
	if validationKeyPem == nil {
		return fmt.Errorf("failed to decode validation key")
	}
	validationKey, err := x509.ParsePKIXPublicKey(validationKeyPem.Bytes)
	if err != nil {
		return err
	}
	c.ValidationKey = validationKey

	previousValidationKeyPem, _ := pem.Decode([]byte(escapeNewLines(c.PreviousValidationKeyString)))
	if previousValidationKeyPem == nil {
		return fmt.Errorf("failed to decode previous validation key")
	}
	previousValidationKey, err := x509.ParsePKIXPublicKey(previousValidationKeyPem.Bytes)
	if err != nil {
		return err
	}
	c.PreviousValidationKey = previousValidationKey
	return nil
}

func escapeNewLines(str string) string {
	return strings.ReplaceAll(str, `\n`, "\n")
}
