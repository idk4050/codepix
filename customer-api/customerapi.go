package customerapi

import (
	"codepix/customer-api/config"
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type CustomerAPI struct {
	logger logr.Logger
	config config.Config
}

func New(ctx context.Context, loggerImpl *zap.Logger, config config.Config) (*CustomerAPI, error) {
	logger := zapr.NewLogger(loggerImpl.WithOptions(
		zap.AddStacktrace(zapcore.DPanicLevel),
		zap.WithCaller(false),
	))

	customerAPI := &CustomerAPI{
		logger: logger,
		config: config,
	}
	return customerAPI, nil
}

func (api CustomerAPI) Start(ctx context.Context) error {
	api.logger.Info("starting customer API")

	api.logger.Info("customer API started")
	return nil
}

func (api CustomerAPI) Stop() error {
	api.logger.Info("stopping customer API")

	api.logger.Info("customer API stopped")
	return nil
}
