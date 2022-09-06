package examplebankapi

import (
	"codepix/example-bank-api/adapters/databaseclient"
	"codepix/example-bank-api/adapters/messagequeue"
	"codepix/example-bank-api/config"
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ExampleBankAPI struct {
	logger       logr.Logger
	config       config.Config
	database     *databaseclient.Database
	messageQueue *messagequeue.MessageQueue
}

func New(ctx context.Context, loggerImpl *zap.Logger, config config.Config) (*ExampleBankAPI, error) {
	logger := zapr.NewLogger(loggerImpl.WithOptions(
		zap.AddStacktrace(zapcore.DPanicLevel),
		zap.WithCaller(false),
	))

	database, err := databaseclient.Open(config, logger)
	if err != nil {
		return nil, err
	}
	messageQueue, err := messagequeue.Open(config, logger)
	if err != nil {
		return nil, err
	}
	api := &ExampleBankAPI{
		logger:       logger,
		config:       config,
		database:     database,
		messageQueue: messageQueue,
	}
	return api, nil
}

func (api ExampleBankAPI) Start(ctx context.Context) error {
	api.logger.Info("starting Example Bank API")

	err := api.database.AutoMigrate()
	if err != nil {
		return err
	}
	api.logger.Info("Example Bank API started")
	return nil
}

func (api ExampleBankAPI) Stop() error {
	api.logger.Info("stopping Example Bank API")

	err := api.database.Close()
	if err != nil {
		return err
	}
	err = api.messageQueue.Close()
	if err != nil {
		return err
	}
	api.logger.Info("Example Bank API stopped")
	return nil
}
