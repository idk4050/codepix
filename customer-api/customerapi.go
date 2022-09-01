package customerapi

import (
	"codepix/customer-api/adapters/databaseclient"
	"codepix/customer-api/adapters/outbox"
	"codepix/customer-api/config"
	"codepix/customer-api/lib/outboxes"
	"codepix/customer-api/lib/publishers"
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
	outbox   outboxes.Outbox
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
	publishers := map[outboxes.Namespace]publishers.Publisher{}
	outbox, err := outbox.Open(config, logger, publishers)
	if err != nil {
		return nil, err
	}

	customerAPI := &CustomerAPI{
		config:   config,
		logger:   logger,
		database: database,
		outbox:   outbox,
	}
	return customerAPI, nil
}

func (api CustomerAPI) Start(ctx context.Context) error {
	api.logger.Info("starting customer API")

	err := api.database.AutoMigrate()
	if err != nil {
		return err
	}
	err = api.outbox.AutoMigrate()
	if err != nil {
		return err
	}
	go api.outbox.Start(ctx)

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
