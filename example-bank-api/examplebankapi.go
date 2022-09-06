package examplebankapi

import (
	"codepix/example-bank-api/config"
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ExampleBankAPI struct {
	logger logr.Logger
	config config.Config
}

func New(ctx context.Context, loggerImpl *zap.Logger, config config.Config) (*ExampleBankAPI, error) {
	logger := zapr.NewLogger(loggerImpl.WithOptions(
		zap.AddStacktrace(zapcore.DPanicLevel),
		zap.WithCaller(false),
	))

	api := &ExampleBankAPI{
		logger: logger,
		config: config,
	}
	return api, nil
}

func (api ExampleBankAPI) Start(ctx context.Context) error {
	api.logger.Info("starting Example Bank API")

	api.logger.Info("Example Bank API started")
	return nil
}

func (api ExampleBankAPI) Stop() error {
	api.logger.Info("stopping Example Bank API")

	api.logger.Info("Example Bank API stopped")
	return nil
}
