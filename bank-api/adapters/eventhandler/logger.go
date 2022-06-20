package eventhandler

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/looplab/eventhorizon"
)

func Logger(logger logr.Logger, handler eventhorizon.EventHandler) eventhorizon.EventHandler {
	return eventhorizon.EventHandlerFunc(func(ctx context.Context, event eventhorizon.Event) error {
		kvs := []any{
			"id", event.AggregateID(),
			"type", event.EventType(),
		}
		if err := handler.HandleEvent(ctx, event); err != nil {
			logger.Error(err, "event failed", kvs...)
			return err
		}
		logger.Info("event handled", kvs...)
		return nil
	})
}
