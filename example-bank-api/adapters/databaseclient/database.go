package databaseclient

import (
	"fmt"

	"codepix/example-bank-api/config"

	"github.com/go-logr/logr"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	*gorm.DB
	config config.Config
	logger logr.Logger
}

func Open(config config.Config, logger logr.Logger) (*Database, error) {
	cfg := config.Database
	logger = logger.WithName("database")

	gormConfig := gorm.Config{
		Logger: NewLogger(logger),
	}

	DSN := cfg.ConnectionString
	if DSN == "" {
		switch cfg.Dialect {
		case "postgres":
			DSN = fmt.Sprintf(
				"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
				cfg.Host, cfg.Port, cfg.Name, cfg.User, cfg.Password, cfg.SSLMode,
			)
		}
	}

	var client *gorm.DB
	var err error
	switch cfg.Dialect {
	case "postgres":
		client, err = gorm.Open(postgres.New(postgres.Config{
			DSN: DSN,
		}), &gormConfig)
	case "sqlite":
		client, err = gorm.Open(sqlite.Open(DSN), &gormConfig)
	default:
		return nil, fmt.Errorf("open database: invalid dialect %s", cfg.Dialect)
	}
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	logger.Info("database opened")

	database := &Database{
		DB:     client,
		config: config,
		logger: logger,
	}
	return database, nil
}

func (db *Database) AutoMigrate(models ...interface{}) error {
	if db.config.Database.AutoMigrate {
		err := db.DB.AutoMigrate(models...)

		if err != nil {
			db.logger.Error(err, "database migration failed")
			return err
		}
		db.logger.Info("database migrated")
	}
	return nil
}

func (db *Database) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		db.logger.Error(err, "database failed to close")
		return err
	}
	err = sqlDB.Close()
	if err != nil {
		db.logger.Error(err, "database failed to close")
		return err
	}
	db.logger.Info("database closed")
	return nil
}
