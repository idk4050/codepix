package eventbus

import (
	"codepix/bank-api/adapters/eventjson"
	"context"

	"github.com/go-redis/redis/v9"
	"github.com/looplab/eventhorizon"
)

type Writer struct {
	Client  *redis.Client
	streams func(eventhorizon.Event) []string
}

func (Writer) HandlerType() eventhorizon.EventHandlerType { return "eventbus" }

func (w Writer) HandleEvent(ctx context.Context, event eventhorizon.Event) error {
	eventJson, err := eventjson.Marshal(event)
	if err != nil {
		return err
	}
	pipe := w.Client.Pipeline()
	for _, stream := range w.streams(event) {
		args := &redis.XAddArgs{
			Stream: stream,
			Values: []string{eventKey, string(eventJson)},
		}
		pipe.XAdd(ctx, args)
	}
	_, err = pipe.Exec(ctx)
	return err
}
