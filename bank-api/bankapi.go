package bankapi

import (
	"codepix/bank-api/adapters/databaseclient"
	"codepix/bank-api/adapters/eventstore"
	"codepix/bank-api/adapters/projectionclient"
	"codepix/bank-api/config"
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

type BankAPI struct {
	config     config.Config
	logger     logr.Logger
	database   *gorm.DB
	eventStore *eventstore.EventStore
	projection *projectionclient.StoreProjection
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
	eventStore, err := eventstore.Open(config, logger)
	if err != nil {
		return nil, err
	}
	projection, err := projectionclient.Open(config, logger, eventStore.Outbox)
	if err != nil {
		return nil, err
	}
	bankAPI := &BankAPI{
		config:     config,
		logger:     logger,
		database:   database,
		eventStore: eventStore,
		projection: projection,
	}
	return bankAPI, nil
}

func (api BankAPI) Start(ctx context.Context) error {
	err := api.database.AutoMigrate()
	if err != nil {
		return err
	}

	api.eventStore.Outbox.Start()
	return nil
}

func (api BankAPI) Stop() error {
	err := api.eventStore.Outbox.Close()
	if err != nil {
		api.logger.Error(err, "event store outbox failed to close")
		return err
	}
	return nil
}
