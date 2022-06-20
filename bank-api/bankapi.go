package bankapi

import (
	"codepix/bank-api/adapters/commandbus"
	"codepix/bank-api/adapters/databaseclient"
	"codepix/bank-api/adapters/eventbus"
	"codepix/bank-api/adapters/eventstore"
	"codepix/bank-api/adapters/projectionclient"
	"codepix/bank-api/adapters/rpc"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/bank/auth"
	"codepix/bank-api/config"
	pixkeydatabase "codepix/bank-api/pixkey/repository/database"
	pixkeyservice "codepix/bank-api/pixkey/service"
	"context"
	"errors"
	"net"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/commandhandler/bus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type BankAPI struct {
	logger     logr.Logger
	config     config.Config
	database   *databaseclient.Database
	eventStore *eventstore.EventStore
	projection *projectionclient.StoreProjection
	eventBus   *eventbus.EventBus
	commandBus eventhorizon.CommandHandler
	server     *grpc.Server
}

func New(ctx context.Context, loggerImpl *zap.Logger, config config.Config) (*BankAPI, error) {
	logger := zapr.NewLogger(loggerImpl.WithOptions(
		zap.AddStacktrace(zapcore.DPanicLevel),
		zap.WithCaller(false),
	))
	panicLogger := zapr.NewLogger(loggerImpl.WithOptions(
		zap.AddStacktrace(zapcore.DebugLevel),
		zap.AddCallerSkip(3),
		zap.Fields(
			zap.StackSkip("stacktrace", 3),
		),
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
	validator, err := validator.New()
	if err != nil {
		return nil, err
	}
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			rpc.UnaryPanicHandler(panicLogger),
			rpc.UnaryLogger(logger),
			auth.UnaryTokenValidator(config),
			rpc.UnaryValidator(validator),
		),
		grpc.ChainStreamInterceptor(
			rpc.StreamPanicHandler(panicLogger),
			rpc.StreamLogger(logger),
			auth.StreamTokenValidator(config),
			rpc.StreamValidator(validator),
		),
	)

	pixKeyRepository := &pixkeydatabase.Database{Database: database}
	err = pixkeyservice.Register(server, validator, pixKeyRepository)
	if err != nil {
		return nil, err
	}

	reflection.Register(server)

	bankAPI := &BankAPI{
		logger:     logger,
		config:     config,
		database:   database,
		eventStore: eventStore,
		projection: projection,
		eventBus:   eventBus,
		commandBus: commandBus,
		server:     server,
	}
	return bankAPI, nil
}

func (api BankAPI) Start(ctx context.Context) error {
	api.logger.Info("starting bank API")

	err := api.database.AutoMigrate(
		&pixkeydatabase.PixKey{},
	)
	if err != nil {
		return err
	}
	err = api.eventStore.Start()
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", "0.0.0.0:"+api.config.RPC.Port)
	if err != nil {
		return err
	}
	grpcLogger := api.logger.WithName("grpc")
	grpcLogger.Info("grpc server listening on port " + api.config.RPC.Port)
	go func() {
		err := api.server.Serve(listener)
		switch {
		case err == nil:
			return
		case errors.Is(err, grpc.ErrServerStopped):
			grpcLogger.Info("grpc server stopped")
		default:
			grpcLogger.Error(err, "grpc server failed to serve")
		}
	}()

	api.logger.Info("bank API started")
	return nil
}

func (api BankAPI) Stop() error {
	api.logger.Info("stopping bank API")

	api.server.Stop()
	api.logger.WithName("grpc").Info("grpc server stopped")

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
