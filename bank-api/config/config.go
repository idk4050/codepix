package config

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/caarlos0/env"
	"github.com/subosito/gotenv"
)

type Config struct {
	Database        database
	EventStore      eventStore
	StoreProjection storeProjection
	RPC             rpc
	BankAuth        bankAuth
}

func New() (*Config, error) {
	c := &Config{
		Database:        database{},
		EventStore:      eventStore{},
		StoreProjection: storeProjection{},
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
	if c.EventStore == (eventStore{}) {
		return nil, errors.New("failed to load event store config")
	}
	env.Parse(&c.StoreProjection)
	if c.StoreProjection == (storeProjection{}) {
		return nil, errors.New("failed to load store projection config")
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
	PasswordFromFile string `env:"DB_PASSWORD_FILE,file"`
	SSLMode          string `env:"DB_SSLMODE"`
}

type eventStore struct {
	InMemory           bool   `env:"ES_IN_MEMORY"`
	InMemoryBinaryPath string `env:"ES_IN_MEMORY_BINARY_PATH"`
	Host               string `env:"ES_HOST"`
	Port               string `env:"ES_PORT"`
	Name               string `env:"ES_NAME"`
	User               string `env:"ES_USER"`
	Password           string `env:"ES_PASSWORD"`
	PasswordFromFile   string `env:"ES_PASSWORD_FILE,file"`
	ReplicaSetName     string `env:"ES_REPLICA_SET_NAME"`
}

type storeProjection struct {
	InMemory           bool   `env:"SP_IN_MEMORY"`
	InMemoryBinaryPath string `env:"SP_IN_MEMORY_BINARY_PATH"`
	Host               string `env:"SP_HOST"`
	Port               string `env:"SP_PORT"`
	Name               string `env:"SP_NAME"`
	User               string `env:"SP_USER"`
	Password           string `env:"SP_PASSWORD"`
	PasswordFromFile   string `env:"SP_PASSWORD_FILE,file"`
	ReplicaSetName     string `env:"SP_REPLICA_SET_NAME"`
}

type rpc struct {
	Port string `env:"RPC_PORT"`
}

type bankAuth struct {
	MetadataKey                 string `env:"BANK_AUTH_METADATA_KEY"`
	ValidationKey               any
	ValidationKeyString         string `env:"BANK_AUTH_VALIDATION_KEY"`
	PreviousValidationKey       any
	PreviousValidationKeyString string `env:"BANK_AUTH_PREVIOUS_VALIDATION_KEY"`
}

func (c *bankAuth) build() error {
	validationKeyPem, _ := pem.Decode([]byte(c.ValidationKeyString))
	validationKey, err := x509.ParsePKIXPublicKey(validationKeyPem.Bytes)
	if err != nil {
		return err
	}
	c.ValidationKey = validationKey

	previousValidationKeyPem, _ := pem.Decode([]byte(c.PreviousValidationKeyString))
	previousValidationKey, err := x509.ParsePKIXPublicKey(previousValidationKeyPem.Bytes)
	if err != nil {
		return err
	}
	c.PreviousValidationKey = previousValidationKey
	return nil
}
