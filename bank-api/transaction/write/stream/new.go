package stream

import (
	"bytes"
	"codepix/bank-api/adapters/eventbus"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/lib/validation"
	pixkeyrepository "codepix/bank-api/pixkey/repository"
	proto "codepix/bank-api/proto/codepix/transaction/write"
	"codepix/bank-api/transaction"
	"codepix/bank-api/transaction/write"

	"github.com/go-logr/logr"
	"github.com/looplab/eventhorizon"
	"google.golang.org/grpc"
)

func Register(logger logr.Logger, server *grpc.Server, val *validation.Validator,
	commandHandler eventhorizon.CommandHandler, pixKeyRepository pixkeyrepository.Repository,
) error {
	err := validator.LoadTranslationFile(val, bytes.NewReader(write.Translations),
		proto.StartRequest{},
		proto.FailRequest{},
	)
	if err != nil {
		return err
	}
	stream := &Stream{
		Logger:           logger.WithName("commandstream"),
		CommandHandler:   commandHandler,
		PixKeyRepository: pixKeyRepository,
	}
	proto.RegisterStreamServer(server, stream)
	return nil
}

func SetupWriters(eventBus *eventbus.EventBus) error {
	err := eventBus.SetupWriter(transaction.StartedEvent, func(event eventhorizon.Event) []string {
		started := event.Data().(*transaction.TransactionStarted)
		return []string{
			transaction.StartedStream(started.ReceiverBank),
		}
	})
	if err != nil {
		return err
	}
	err = eventBus.SetupWriter(transaction.ConfirmedEvent, func(event eventhorizon.Event) []string {
		confirmed := event.Data().(*transaction.TransactionConfirmed)
		return []string{
			transaction.ConfirmedStream(confirmed.SenderBank),
		}
	})
	if err != nil {
		return err
	}
	err = eventBus.SetupWriter(transaction.CompletedEvent, func(event eventhorizon.Event) []string {
		completed := event.Data().(*transaction.TransactionCompleted)
		return []string{
			transaction.CompletedStream(completed.ReceiverBank),
		}
	})
	if err != nil {
		return err
	}
	err = eventBus.SetupWriter(transaction.FailedEvent, func(event eventhorizon.Event) []string {
		failed := event.Data().(*transaction.TransactionFailed)
		return []string{
			transaction.FailedStream(failed.SenderBank),
			transaction.FailedStream(failed.ReceiverBank),
		}
	})
	if err != nil {
		return err
	}
	return nil
}
