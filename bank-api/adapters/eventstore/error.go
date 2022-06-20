package eventstore

import (
	"codepix/bank-api/lib/eventrepositories"
	"errors"

	"github.com/looplab/eventhorizon"
)

func MapError(err error) error {
	switch err := err.(type) {
	case *eventhorizon.EventStoreError:
		aggregate := eventrepositories.ErrorAggregate{
			Type:    string(err.AggregateType),
			ID:      err.AggregateID,
			Version: uint(err.AggregateVersion),
		}
		switch {
		case errors.Is(err.Err, eventhorizon.ErrAggregateNotFound):
			return eventrepositories.NewNotFoundError(aggregate)
		case errors.Is(err.Err, eventhorizon.ErrEventConflictFromOtherSave):
			return eventrepositories.NewVersionConflictError(aggregate)
		default:
			return eventrepositories.NewInternalError(aggregate, string(err.Op), err.Error())
		}
	default:
		return err
	}
}
