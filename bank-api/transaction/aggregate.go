package transaction

import (
	"codepix/bank-api/lib/aggregates"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/aggregatestore/events"
)

const AggregateType = eventhorizon.AggregateType("transaction")

func init() {
	eventhorizon.RegisterAggregate(func(ID uuid.UUID) eventhorizon.Aggregate { return New(ID) })
}

type Aggregate struct {
	*events.AggregateBase
	Transaction *Transaction
}

func New(ID uuid.UUID) *Aggregate {
	return &Aggregate{
		AggregateBase: events.NewAggregateBase(AggregateType, ID),
		Transaction:   &Transaction{},
	}
}

func (ag Aggregate) HandleCommand(ctx context.Context, cmd eventhorizon.Command) error {
	if cmd, ok := cmd.(Command); ok {
		eventData, eventType, err := cmd.ToEvent(ag)
		if err != nil {
			return aggregates.NewInvariantViolation(err.Error())
		}
		ag.AppendEvent(eventType, eventData, time.Now())
		return nil
	}
	return fmt.Errorf("unknown command type %s", cmd.CommandType())
}

func (ag *Aggregate) ApplyEvent(ctx context.Context, event eventhorizon.Event) error {
	if eventData, ok := event.Data().(Event); ok {
		eventData.Apply(ag.Transaction)
		return nil
	}
	return fmt.Errorf("unknown event type %s", event.EventType())
}
