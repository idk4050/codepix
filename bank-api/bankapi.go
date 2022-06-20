package bankapi

import (
	"codepix/bank-api/config"
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type BankAPI struct {
	logger logr.Logger
	config config.Config
}

func New(ctx context.Context, loggerImpl *zap.Logger, config config.Config) (*BankAPI, error) {
	logger := zapr.NewLogger(loggerImpl.WithOptions(
		zap.AddStacktrace(zapcore.DPanicLevel),
		zap.WithCaller(false),
	))
	bankAPI := &BankAPI{
		logger: logger,
		config: config,
	}
	return bankAPI, nil
}

func (api BankAPI) Start(ctx context.Context) error {
	api.logger.Info("starting bank API")
	api.logger.Info("bank API started")
	return nil
}

func (api BankAPI) Stop() error {
	api.logger.Info("stopping bank API")
	api.logger.Info("bank API stopped")
	return nil
}
