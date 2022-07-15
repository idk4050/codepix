package bankapi

import (
	"codepix/bank-api/adapters/commandbus"
	"codepix/bank-api/adapters/databaseclient"
	"codepix/bank-api/adapters/eventbus"
	"codepix/bank-api/adapters/eventstore"
	"codepix/bank-api/adapters/projectionclient"
	"codepix/bank-api/config"
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/commandhandler/bus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type BankAPI struct {
	logger     logr.Logger
	config     config.Config
	database   *databaseclient.Database
	eventStore *eventstore.EventStore
	projection *projectionclient.StoreProjection
	eventBus   *eventbus.EventBus
	commandBus eventhorizon.CommandHandler
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
	eventStore, err := eventstore.Open(ctx, config, logger)
	if err != nil {
		return nil, err
	}
	projection, err := projectionclient.Open(ctx, config, logger, eventStore.Outbox)
	if err != nil {
		return nil, err
	}
	eventBus, err := eventbus.Open(ctx, config, logger, eventStore.Outbox)
	if err != nil {
		return nil, err
	}
	commandBusHandler := bus.NewCommandHandler()
	commandBus := eventhorizon.UseCommandHandlerMiddleware(commandBusHandler,
		commandbus.Logger(logger),
	)
	bankAPI := &BankAPI{
		logger:     logger,
		config:     config,
		database:   database,
		eventStore: eventStore,
		projection: projection,
		eventBus:   eventBus,
		commandBus: commandBus,
	}
	return bankAPI, nil
}

func (api BankAPI) Start(ctx context.Context) error {
	api.logger.Info("starting bank API")

	err := api.database.AutoMigrate()
	if err != nil {
		return err
	}
	err = api.eventStore.Start()
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
	err = api.eventStore.Close()
	if err != nil {
		return err
	}
	err = api.projection.Close()
	if err != nil {
		return err
	}
	err = api.eventBus.Close()
	if err != nil {
		return err
	}
	api.logger.Info("bank API stopped")
	return nil
}
