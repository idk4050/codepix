package stream

import (
	accountrepository "codepix/bank-api/account/repository"
	"codepix/bank-api/lib/validation"
	"codepix/bank-api/transaction/stream/proto"

	"github.com/looplab/eventhorizon"
	"google.golang.org/grpc"
)

func Register(server *grpc.Server, val *validation.Validator, commandHandler eventhorizon.CommandHandler,
	outbox eventhorizon.Outbox, accountRepository accountrepository.Repository,
) error {
	startedBus := StartedBus{Outbox: outbox}
	confirmedBus := ConfirmedBus{Outbox: outbox}
	stream := &Stream{
		StartedBus:        startedBus,
		ConfirmedBus:      confirmedBus,
		CommandHandler:    commandHandler,
		AccountRepository: accountRepository,
	}
	proto.RegisterTransactionStreamServer(server, stream)
	return nil
}
