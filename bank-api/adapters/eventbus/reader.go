package eventbus

import (
	"codepix/bank-api/adapters/eventjson"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/looplab/eventhorizon"
)

type Reader struct {
	Client        *redis.Client
	BlockDuration time.Duration
	MaxPendingAge time.Duration
}

func (r Reader) CreateGroup(ctx context.Context, stream, group string) error {
	response, err := r.Client.XGroupCreateMkStream(ctx, stream, group, "0").Result()
	if err != nil {
		if !strings.HasPrefix(err.Error(), "BUSYGROUP") {
			return fmt.Errorf("create consumer group: %w", err)
		}
	} else if response != "OK" {
		err := fmt.Errorf("got %s response", response)
		return fmt.Errorf("create consumer group: %w", err)
	}
	return nil
}

func (r Reader) Ack(ctx context.Context, stream, group string, messageIDs []string) error {
	_, err := r.Client.XAck(ctx, stream, group, messageIDs...).Result()
	if err != nil {
		return fmt.Errorf("ack messages: %w", err)
	}
	return nil
}

func (r Reader) Consume(ctx context.Context, stream, group, consumer string,
) ([]eventhorizon.Event, []string, error) {
	var messages []redis.XMessage

	pending, _, err := r.Client.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   stream,
		Group:    group,
		Consumer: consumer,
		Start:    "0",
		MinIdle:  r.MaxPendingAge,
	}).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("consume: get pending events: %w", err)
	}
	if len(pending) > 0 {
		messages = pending
	} else {
		streamSlices, err := r.Client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Streams:  []string{stream, ">"},
			Group:    group,
			Consumer: consumer,
			Block:    r.BlockDuration,
		}).Result()
		if err == redis.Nil {
			return []eventhorizon.Event{}, []string{}, nil
		}
		if err != nil {
			return nil, nil, fmt.Errorf("consume: get events: %w", err)
		}
		messages = streamSlices[0].Messages
	}

	events := []eventhorizon.Event{}
	messageIDs := []string{}
	for _, message := range messages {
		eventJson := message.Values[eventKey].(string)
		event, err := eventjson.Unmarshal([]byte(eventJson))
		if err != nil {
			return nil, nil, fmt.Errorf("consume: unmarshal event: %w", err)
		}
		events = append(events, event)
		messageIDs = append(messageIDs, message.ID)
	}
	return events, messageIDs, nil
}
