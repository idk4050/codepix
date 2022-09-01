package customerapi

import (
	"codepix/customer-api/adapters/databaseclient"
	"codepix/customer-api/config"
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type CustomerAPI struct {
	logger   logr.Logger
	config   config.Config
	database *databaseclient.Database
}

func New(ctx context.Context, loggerImpl *zap.Logger, config config.Config) (*CustomerAPI, error) {
	logger := zapr.NewLogger(loggerImpl.WithOptions(
		zap.AddStacktrace(zapcore.DPanicLevel),
		zap.WithCaller(false),
	))
	database, err := databaseclient.Open(config, logger)
	if err != nil {
		return nil, err
	}

	customerAPI := &CustomerAPI{
		config:   config,
		logger:   logger,
		database: database,
	}
	return customerAPI, nil
}

func (api CustomerAPI) Start(ctx context.Context) error {
	api.logger.Info("starting customer API")

	err := api.database.AutoMigrate()
	if err != nil {
		return err
	}

	api.logger.Info("customer API started")
	return nil
}

func (api CustomerAPI) Stop() error {
	api.logger.Info("stopping customer API")

	err := api.database.Close()
	if err != nil {
		return err
	}
	api.logger.Info("customer API stopped")
	return nil
}
