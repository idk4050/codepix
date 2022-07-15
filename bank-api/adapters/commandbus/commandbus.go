package commandbus

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/looplab/eventhorizon"
)

func Logger(logger logr.Logger) func(eventhorizon.CommandHandler) eventhorizon.CommandHandler {
	log := logger.WithName("commandbus")

	return func(h eventhorizon.CommandHandler) eventhorizon.CommandHandler {
		return eventhorizon.CommandHandlerFunc(func(ctx context.Context, cmd eventhorizon.Command) error {
			kvs := []any{
				"aggregate", cmd.AggregateType(),
				"type", cmd.CommandType(),
				"id", cmd.AggregateID(),
			}
			if err := h.HandleCommand(ctx, cmd); err != nil {
				log.Error(err, "command not handled", kvs...)
				return err
			}
			log.Info("command handled", kvs...)
			return nil
		})
	}
}
