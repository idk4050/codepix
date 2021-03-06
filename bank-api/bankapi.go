package bankapi

import (
	accountusecase "codepix/bank-api/account/interactor/usecase"
	accountdatabase "codepix/bank-api/account/repository/database"
	accountservice "codepix/bank-api/account/service"
	"codepix/bank-api/adapters/commandbus"
	"codepix/bank-api/adapters/databaseclient"
	"codepix/bank-api/adapters/eventstore"
	"codepix/bank-api/adapters/projectionclient"
	"codepix/bank-api/adapters/rpc"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/bank/auth"
	"codepix/bank-api/config"
	pixkeyusecase "codepix/bank-api/pixkey/interactor/usecase"
	pixkeydatabase "codepix/bank-api/pixkey/repository/database"
	pixkeyservice "codepix/bank-api/pixkey/service"
	transactioncommandhandler "codepix/bank-api/transaction/commandhandler"
	transactionprojection "codepix/bank-api/transaction/readrepository/projection"
	transactionservice "codepix/bank-api/transaction/service"
	transactionstream "codepix/bank-api/transaction/stream"
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
	"gorm.io/gorm"
)

type BankAPI struct {
	config     config.Config
	logger     logr.Logger
	database   *gorm.DB
	eventStore *eventstore.EventStore
	projection *projectionclient.StoreProjection
	commandBus eventhorizon.CommandHandler
	server     *grpc.Server
	listener   net.Listener
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
			rpc.UnaryLogger(logger),
			rpc.UnaryValidator(validator),
			auth.UnaryTokenValidator(config),
		),
		grpc.ChainStreamInterceptor(
			rpc.StreamLogger(logger),
			rpc.StreamValidator(validator),
			auth.StreamTokenValidator(config),
		),
	)

	accountRepository := &accountdatabase.Database{DB: database}
	accountInteractor := &accountusecase.Usecase{Repository: accountRepository}
	err = accountservice.Register(server, validator, accountInteractor, accountRepository)
	if err != nil {
		return nil, err
	}

	pixKeyRepository := &pixkeydatabase.Database{DB: database}
	pixKeyInteractor := &pixkeyusecase.Usecase{
		Repository:        pixKeyRepository,
		AccountRepository: accountRepository,
	}
	err = pixkeyservice.Register(server, validator, pixKeyInteractor, pixKeyRepository, accountRepository)
	if err != nil {
		return nil, err
	}

	err = transactioncommandhandler.Setup(eventStore, commandBusHandler)
	if err != nil {
		return nil, err
	}
	transactionRepository, err := transactionprojection.New(projection)
	if err != nil {
		return nil, err
	}
	err = transactionservice.Register(server, validator, commandBus,
		transactionRepository, accountRepository, pixKeyRepository)
	if err != nil {
		return nil, err
	}
	err = transactionstream.Register(server, validator, commandBus, eventStore.Outbox,
		accountRepository)
	if err != nil {
		return nil, err
	}

	reflection.Register(server)
	bankAPI := &BankAPI{
		config:     config,
		logger:     logger,
		database:   database,
		eventStore: eventStore,
		projection: projection,
		commandBus: commandBus,
		server:     server,
	}
	return bankAPI, nil
}

func (api BankAPI) Start(ctx context.Context) error {
	err := api.database.AutoMigrate(
		&accountdatabase.Account{},
		&pixkeydatabase.PixKey{},
	)
	if err != nil {
		return err
	}

	api.eventStore.Outbox.Start()

	listener, err := net.Listen("tcp", "0.0.0.0:"+api.config.RPC.Port)
	if err != nil {
		return err
	}
	api.listener = listener
	api.logger.Info("server listening on port " + api.config.RPC.Port)

	go func() {
		err := api.server.Serve(listener)
		switch {
		case err == nil:
			return
		case errors.Is(err, grpc.ErrServerStopped):
			api.logger.Info("server stopped")
		default:
			api.logger.Error(err, "server failed to serve")
		}
	}()
	return nil
}

func (api BankAPI) Stop() error {
	api.server.GracefulStop()
	api.logger.Info("server stopped")

	err := api.listener.Close()
	if err != nil {
		api.logger.Error(err, "server listener failed to close")
		return err
	}
	api.logger.Info("server listener closed")

	err = api.eventStore.Outbox.Close()
	if err != nil {
		api.logger.Error(err, "event store outbox failed to close")
		return err
	}
	return nil
}
