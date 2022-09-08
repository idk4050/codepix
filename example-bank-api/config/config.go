package config

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/golang-jwt/jwt"
	"github.com/subosito/gotenv"
)

type Config struct {
	Database     database
	MessageQueue messageQueue
	HTTP         http
	UserAuth     userAuth
	PixAPI       pixAPI
}

func New() (*Config, error) {
	c := &Config{
		Database:     database{},
		MessageQueue: messageQueue{},
		HTTP:         http{},
		UserAuth:     userAuth{},
		PixAPI:       pixAPI{},
	}
	err := loadEnvFileIfAvailable()
	if err != nil {
		return nil, err
	}
	env.Parse(&c.Database)
	if c.Database == (database{}) {
		return nil, errors.New("failed to load database config")
	}
	env.Parse(&c.MessageQueue)
	if c.MessageQueue == (messageQueue{}) {
		return nil, errors.New("failed to load message queue config")
	}
	env.Parse(&c.HTTP)
	if c.HTTP == (http{}) {
		return nil, errors.New("failed to load HTTP config")
	}
	env.Parse(&c.UserAuth)
	if c.UserAuth == (userAuth{}) {
		return nil, errors.New("failed to load user auth config")
	}
	err = c.UserAuth.build()
	if err != nil {
		return nil, fmt.Errorf("failed to build user auth config: %w", err)
	}
	env.Parse(&c.PixAPI)
	if c.PixAPI == (pixAPI{}) {
		return nil, errors.New("failed to load Pix API config")
	}
	err = c.PixAPI.build()
	if err != nil {
		return nil, fmt.Errorf("failed to build Pix API config: %w", err)
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

type messageQueue struct {
	InMemory bool   `env:"MQ_IN_MEMORY"`
	Host     string `env:"MQ_HOST"`
	Port     string `env:"MQ_PORT"`
	Name     string `env:"MQ_NAME"`
	User     string `env:"MQ_USER"`
	Password string `env:"MQ_PASSWORD"`
}

type http struct {
	Port string `env:"HTTP_PORT"`
}

type userAuth struct {
	CookieName                  string `env:"USER_AUTH_COOKIE_NAME"`
	TimeUntilExpiration         time.Duration
	MinutesUntilExpirationInt   uint `env:"USER_AUTH_MINUTES_UNTIL_EXPIRATION"`
	SigningMethod               jwt.SigningMethod
	SigningKey                  any
	SigningKeyString            string `env:"USER_AUTH_SIGNING_KEY"`
	ValidationKey               any
	ValidationKeyString         string `env:"USER_AUTH_VALIDATION_KEY"`
	PreviousValidationKey       any
	PreviousValidationKeyString string `env:"USER_AUTH_PREVIOUS_VALIDATION_KEY"`
}

func (c *userAuth) build() error {
	signingKeyPem, _ := pem.Decode([]byte(escapeNewLines(c.SigningKeyString)))
	if signingKeyPem == nil {
		return fmt.Errorf("failed to decode signing key")
	}
	signingKey, err := x509.ParsePKCS8PrivateKey(signingKeyPem.Bytes)
	if err != nil {
		return fmt.Errorf("invalid signing key: %w", err)
	}
	c.SigningKey = signingKey

	signingMethod := getSigningMethod(signingKey)
	if signingMethod == nil {
		return errors.New("no signing method found for key")
	}
	c.SigningMethod = signingMethod

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

	c.TimeUntilExpiration = time.Minute * time.Duration(c.MinutesUntilExpirationInt)
	return nil
}

func getSigningMethod(key any) jwt.SigningMethod {
	switch key.(type) {
	case *rsa.PrivateKey:
		return jwt.SigningMethodRS512
	case *ecdsa.PrivateKey:
		return jwt.SigningMethodES512
	case *ed25519.PrivateKey:
		return jwt.SigningMethodEdDSA
	default:
		return nil
	}
}

func escapeNewLines(str string) string {
	return strings.ReplaceAll(str, `\n`, "\n")
}

type pixAPI struct {
	Address       string
	Host          string `env:"PIX_API_HOST"`
	Port          string `env:"PIX_API_PORT"`
	TokenEndpoint string `env:"PIX_API_TOKEN_ENDPOINT"`
	APIKey        string `env:"PIX_API_KEY"`
}

func (c *pixAPI) build() error {
	ip, err := net.LookupIP(c.Host)
	if err != nil {
		return err
	}
	c.Address = fmt.Sprintf("%s:%s", ip[0], c.Port)
	return nil
}
