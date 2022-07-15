package commandbus

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/looplab/eventhorizon"
)

func Logger(logger logr.Logger) func(eventhorizon.CommandHandler) eventhorizon.CommandHandler {
	logger = logger.WithName("commandbus")

	return func(h eventhorizon.CommandHandler) eventhorizon.CommandHandler {
		return eventhorizon.CommandHandlerFunc(func(ctx context.Context, cmd eventhorizon.Command) error {
			kvs := []any{
				"id", cmd.AggregateID(),
				"type", cmd.CommandType(),
			}
			if err := h.HandleCommand(ctx, cmd); err != nil {
				logger.Error(err, "command failed", kvs...)
				return err
			}
			logger.Info("command handled", kvs...)
			return nil
		})
	}
}
