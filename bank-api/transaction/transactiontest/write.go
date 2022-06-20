package transactiontest

import (
	"codepix/bank-api/adapters/commandbus"
	"codepix/bank-api/adapters/databaseclient"
	"codepix/bank-api/adapters/eventbus"
	"codepix/bank-api/adapters/eventstore"
	"codepix/bank-api/adapters/projectionclient"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/bankapitest"
	"codepix/bank-api/pixkey/pixkeytest"
	pixkeyrepository "codepix/bank-api/pixkey/repository"
	pixkeydatabase "codepix/bank-api/pixkey/repository/database"
	proto "codepix/bank-api/proto/codepix/transaction/write"
	readrepository "codepix/bank-api/transaction/read/repository"
	"codepix/bank-api/transaction/read/repository/projection"
	"codepix/bank-api/transaction/write/commandhandler"
	"codepix/bank-api/transaction/write/service"
	"codepix/bank-api/transaction/write/stream"
	"context"

	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/commandhandler/bus"
)

func CommandHandler() (eventhorizon.CommandHandler, *eventstore.EventStore, TearDown) {
	eventStore, err := eventstore.Open(context.Background(), bankapitest.Config, bankapitest.Logger)
	if err != nil {
		panic(err)
	}
	tearDown := func() {
		err := eventStore.Close()
		if err != nil {
			panic(err)
		}
	}
	commandBusHandler := bus.NewCommandHandler()
	commandBus := eventhorizon.UseCommandHandlerMiddleware(commandBusHandler,
		commandbus.Logger(bankapitest.Logger),
	)
	err = commandhandler.Setup(eventStore, commandBusHandler)
	if err != nil {
		panic(err)
	}
	return commandBus, eventStore, tearDown
}

func WriteStream() (proto.StreamClient, readrepository.Repository,
	pixkeyrepository.Repository, Creator, TearDown) {
	validator, err := validator.New()
	if err != nil {
		panic(err)
	}
	server, client, serve := bankapitest.Server(validator)

	commandHandler, store, storeTearDown := CommandHandler()

	storeProjection, err := projectionclient.Open(context.Background(),
		bankapitest.Config, bankapitest.Logger, store.Outbox)
	if err != nil {
		panic(err)
	}
	readRepository, err := projection.New(storeProjection)
	if err != nil {
		panic(err)
	}

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

	err = stream.Register(bankapitest.Logger, server, validator, commandHandler, pixKeyRepo)
	if err != nil {
		panic(err)
	}
	eventBus, err := eventbus.Open(context.Background(),
		bankapitest.Config, bankapitest.Logger, store.Outbox)
	if err != nil {
		panic(err)
	}
	err = stream.SetupWriters(eventBus)
	if err != nil {
		panic(err)
	}
	store.Start()
	serve()

	receiverIDs := pixkeytest.PixKeyIDs(pixKeyRepo)
	startedIDs := StartedIDs(commandHandler, receiverIDs)
	confirmedIDs := ConfirmedIDs(commandHandler, startedIDs)
	completedIDs := CompletedIDs(commandHandler, confirmedIDs)
	failedIDs := FailedIDs(commandHandler, confirmedIDs)

	creator := Creator{
		ReceiverIDs:  receiverIDs,
		StartedIDs:   startedIDs,
		ConfirmedIDs: confirmedIDs,
		CompletedIDs: completedIDs,
		FailedIDs:    failedIDs,
	}
	tearDown := func() {
		storeTearDown()
		err := eventBus.Close()
		if err != nil {
			panic(err)
		}
		err = storeProjection.Close()
		if err != nil {
			panic(err)
		}
	}
	return proto.NewStreamClient(client), readRepository, pixKeyRepo, creator, tearDown
}

func WriteStreamWithMocks() (proto.StreamClient, *MockCommandHandler, *pixkeytest.MockRepo) {
	validator, err := validator.New()
	if err != nil {
		panic(err)
	}
	server, client, serve := bankapitest.Server(validator)
	commandHandler := new(MockCommandHandler)
	pixKeyRepo := new(pixkeytest.MockRepo)

	err = stream.Register(bankapitest.Logger, server, validator, commandHandler, pixKeyRepo)
	if err != nil {
		panic(err)
	}
	serve()
	return proto.NewStreamClient(client), commandHandler, pixKeyRepo
}

func WriteService() (proto.ServiceClient, readrepository.Repository,
	pixkeyrepository.Repository, Creator, TearDown) {
	validator, err := validator.New()
	if err != nil {
		panic(err)
	}
	server, client, serve := bankapitest.Server(validator)

	commandHandler, store, storeTearDown := CommandHandler()

	storeProjection, err := projectionclient.Open(context.Background(),
		bankapitest.Config, bankapitest.Logger, store.Outbox)
	if err != nil {
		panic(err)
	}
	readRepository, err := projection.New(storeProjection)
	if err != nil {
		panic(err)
	}

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

	err = service.Register(server, validator, commandHandler, pixKeyRepo)
	if err != nil {
		panic(err)
	}
	store.Start()
	serve()

	receiverIDs := pixkeytest.PixKeyIDs(pixKeyRepo)
	startedIDs := StartedIDs(commandHandler, receiverIDs)

	creator := Creator{
		ReceiverIDs: receiverIDs,
		StartedIDs:  startedIDs,
	}
	tearDown := func() {
		storeTearDown()
		err = storeProjection.Close()
		if err != nil {
			panic(err)
		}
	}
	return proto.NewServiceClient(client), readRepository, pixKeyRepo, creator, tearDown
}

func WriteServiceWithMocks() (proto.ServiceClient, *MockCommandHandler, *pixkeytest.MockRepo) {
	validator, err := validator.New()
	if err != nil {
		panic(err)
	}
	server, client, serve := bankapitest.Server(validator)
	commandHandler := new(MockCommandHandler)
	pixKeyRepo := new(pixkeytest.MockRepo)

	err = service.Register(server, validator, commandHandler, pixKeyRepo)
	if err != nil {
		panic(err)
	}
	serve()
	return proto.NewServiceClient(client), commandHandler, pixKeyRepo
}
