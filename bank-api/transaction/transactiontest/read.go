package transactiontest

import (
	"codepix/bank-api/adapters/databaseclient"
	"codepix/bank-api/adapters/eventbus"
	"codepix/bank-api/adapters/jwtclaims"
	"codepix/bank-api/adapters/projectionclient"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/bank/auth"
	"codepix/bank-api/bankapitest"
	"codepix/bank-api/pixkey/pixkeytest"
	pixkeydatabase "codepix/bank-api/pixkey/repository/database"
	proto "codepix/bank-api/proto/codepix/transaction/read"
	"codepix/bank-api/transaction"
	"codepix/bank-api/transaction/read/repository"
	"codepix/bank-api/transaction/read/repository/projection"
	"codepix/bank-api/transaction/read/service"
	"codepix/bank-api/transaction/read/stream"
	"context"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
)

func ReadRepo() (repository.Repository, eventhorizon.CommandHandler, TearDown) {
	commandHandler, store, storeTearDown := CommandHandler()

	projectionClient, err := projectionclient.Open(context.Background(),
		bankapitest.Config, bankapitest.Logger, store.Outbox)
	if err != nil {
		panic(err)
	}
	projection, err := projection.New(projectionClient)
	if err != nil {
		panic(err)
	}
	store.Start()
	tearDown := func() {
		storeTearDown()
		err := projectionClient.Close()
		if err != nil {
			panic(err)
		}
	}
	return projection, commandHandler, tearDown
}

func ReadService() (proto.ServiceClient, repository.Repository, Creator, TearDown) {
	validator, err := validator.New()
	if err != nil {
		panic(err)
	}
	server, client, serve := bankapitest.Server(validator)
	readRepo, commandHandler, storeTearDown := ReadRepo()

	database, err := databaseclient.Open(bankapitest.Config, bankapitest.Logger)
	if err != nil {
		panic(err)
	}
	err = database.AutoMigrate(
		&pixkeydatabase.PixKey{},
	)
	if err != nil {
		panic(err)
	}
	pixKeyRepo := &pixkeydatabase.Database{Database: database}

	err = service.Register(server, readRepo)
	if err != nil {
		panic(err)
	}
	serve()

	receiverIDs := pixkeytest.PixKeyIDs(pixKeyRepo)
	creator := Creator{
		ReceiverIDs: receiverIDs,
		StartedID:   StartedID(commandHandler),
		StartedIDs:  StartedIDs(commandHandler, receiverIDs),
	}
	return proto.NewServiceClient(client), readRepo, creator, storeTearDown
}

func ReadServiceWithMocks() (proto.ServiceClient, *MockCommandHandler, *MockReadRepo,
	*pixkeytest.MockRepo) {
	validator, err := validator.New()
	if err != nil {
		panic(err)
	}
	server, client, serve := bankapitest.Server(validator)
	commandHandler := new(MockCommandHandler)
	projection := new(MockReadRepo)
	pixKeyRepo := new(pixkeytest.MockRepo)

	err = service.Register(server, projection)
	if err != nil {
		panic(err)
	}
	serve()
	return proto.NewServiceClient(client), commandHandler, projection, pixKeyRepo
}

type makeCtx = func(bankID uuid.UUID) context.Context

func ReadStream() (*stream.Stream, makeCtx, eventhorizon.CommandHandler, TearDown) {
	commandHandler, store, storeTearDown := CommandHandler()

	eventBus, err := eventbus.Open(context.Background(), bankapitest.Config, bankapitest.Logger, store.Outbox)
	if err != nil {
		panic(err)
	}
	err = eventBus.SetupWriter(transaction.StartedEvent, func(event eventhorizon.Event) []string {
		started := event.Data().(*transaction.TransactionStarted)
		return []string{transaction.StartedStream(started.ReceiverBank)}
	})
	if err != nil {
		panic(err)
	}
	cfg := bankapitest.Config.Transaction
	busReader, err := eventBus.CreateReader(cfg.BusBlockDuration, cfg.BusMaxPendingAge)
	if err != nil {
		panic(err)
	}
	stream := &stream.Stream{
		Logger:    bankapitest.Logger.WithName("eventstream"),
		BusReader: busReader,
	}
	store.Start()

	makeCtx := func(bankID uuid.UUID) context.Context {
		authClaims := jwt.MapClaims{
			auth.BankIDKey: bankID,
		}
		return jwtclaims.AddClaims(context.Background(), authClaims)
	}
	tearDown := func() {
		storeTearDown()
		err := eventBus.Close()
		if err != nil {
			panic(err)
		}
	}
	return stream, makeCtx, commandHandler, tearDown
}
