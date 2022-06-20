package commandhandler

import (
	"codepix/bank-api/adapters/eventstore"
	"codepix/bank-api/transaction"

	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/aggregatestore/events"
	"github.com/looplab/eventhorizon/commandhandler/aggregate"
	"github.com/looplab/eventhorizon/commandhandler/bus"
)

func Setup(eventStore *eventstore.EventStore, commandBus *bus.CommandHandler) error {
	aggregateStore, err := events.NewAggregateStore(eventStore.Store)
	if err != nil {
		return err
	}
	commandHandler, err := aggregate.NewCommandHandler(transaction.AggregateType, aggregateStore)
	if err != nil {
		return err
	}
	commands := []eventhorizon.CommandType{
		transaction.StartCommand,
		transaction.ConfirmCommand,
		transaction.CompleteCommand,
		transaction.FailCommand,
	}
	for _, cmdType := range commands {
		if err := commandBus.SetHandler(commandHandler, cmdType); err != nil {
			return err
		}
	}
	return nil
}
