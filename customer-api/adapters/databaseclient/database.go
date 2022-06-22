package databaseclient

import (
	"fmt"

	"codepix/customer-api/config"

	"github.com/go-logr/logr"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Open(config config.Config, logger logr.Logger) (*gorm.DB, error) {
	cfg := config.Database

	gormConfig := gorm.Config{
		Logger: NewLogger(logger),
	}

	var client *gorm.DB
	var err error

	password := cfg.Password
	if cfg.PasswordFromFile != "" {
		password = cfg.PasswordFromFile
	}
	switch cfg.Dialect {
	case "postgres":
		DSN := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
			cfg.Host,
			cfg.Port,
			cfg.Name,
			cfg.User,
			password,
			cfg.SSLMode,
		)
		client, err = gorm.Open(postgres.New(postgres.Config{
			DSN: DSN,
		}), &gormConfig)
	case "sqlite":
		DSN := cfg.ConnectionString
		client, err = gorm.Open(sqlite.Open(DSN), &gormConfig)
	default:
		return nil, fmt.Errorf("open database client: invalid dialect %s", cfg.Dialect)
	}

	if err != nil {
		return nil, fmt.Errorf("open database client: %w", err)
	}
	return client, nil
}
