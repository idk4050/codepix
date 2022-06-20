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
func init() {
	eventhorizon.RegisterAggregate(func(ID uuid.UUID) eventhorizon.Aggregate { return New(ID) })
}

func (ag Aggregate) HandleCommand(ctx context.Context, command eventhorizon.Command) error {
	if cmd, ok := command.(Command); ok {
		event, err := cmd.ToEvent(ag)
		if err != nil {
			return &aggregates.InvariantViolation{err}
		}
		ag.AppendEvent(event.Type(), event, time.Now())
		return nil
	}
	return fmt.Errorf("unknown command type %s/%T", command.CommandType(), command)
}

func (ag *Aggregate) ApplyEvent(ctx context.Context, event eventhorizon.Event) error {
	if eventData, ok := event.Data().(Event); ok {
		eventData.Apply(ag.Transaction)
		return nil
	}
	return fmt.Errorf("unknown event type %s/%T", event.EventType(), event.Data())
}
