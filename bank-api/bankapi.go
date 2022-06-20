package bankapi

import (
	"codepix/bank-api/adapters/databaseclient"
	"codepix/bank-api/config"
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type BankAPI struct {
	logger   logr.Logger
	config   config.Config
	database *databaseclient.Database
}

func New(ctx context.Context, loggerImpl *zap.Logger, config config.Config) (*BankAPI, error) {
	logger := zapr.NewLogger(loggerImpl.WithOptions(
		zap.AddStacktrace(zapcore.DPanicLevel),
		zap.WithCaller(false),
	))
	database, err := databaseclient.Open(config, logger)
	if err != nil {
		return nil, err
	}
	bankAPI := &BankAPI{
		logger:   logger,
		config:   config,
		database: database,
	}
	return bankAPI, nil
}

func (api BankAPI) Start(ctx context.Context) error {
	api.logger.Info("starting bank API")

	err := api.database.AutoMigrate()
	if err != nil {
		return err
	}
	api.logger.Info("bank API started")
	return nil
}

func (api BankAPI) Stop() error {
	api.logger.Info("stopping bank API")

	err := api.database.Close()
	if err != nil {
		return err
	}
	api.logger.Info("bank API stopped")
	return nil
}
