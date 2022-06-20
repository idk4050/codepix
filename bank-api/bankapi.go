package bankapi

import (
	"codepix/bank-api/adapters/databaseclient"
	"codepix/bank-api/config"
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

type BankAPI struct {
	config   config.Config
	logger   logr.Logger
	database *gorm.DB
}

func New(config config.Config, loggerImpl *zap.Logger) (*BankAPI, error) {
	logger := zapr.NewLogger(loggerImpl.WithOptions(
		zap.AddStacktrace(zapcore.DPanicLevel),
		zap.WithCaller(false),
	))
	database, err := databaseclient.Open(config, logger)
	if err != nil {
		return nil, err
	}
	bankAPI := &BankAPI{
		config:   config,
		logger:   logger,
		database: database,
	}
	return bankAPI, nil
}

func (api BankAPI) Start(ctx context.Context) error {
	err := api.database.AutoMigrate()
	if err != nil {
		return err
	}
	return nil
}

func (api BankAPI) Stop() error {
	return nil
}
