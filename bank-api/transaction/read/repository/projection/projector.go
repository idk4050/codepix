package projection

import (
	"codepix/bank-api/transaction"
	"codepix/bank-api/transaction/read/repository"
	"context"
	"fmt"

	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/eventhandler/projector"
)

type Projector struct{}

var _ projector.Projector = Projector{}

func (p Projector) ProjectorType() projector.Type {
	return projector.Type(repository.RepositoryType)
}

func (p Projector) Project(ctx context.Context, event eventhorizon.Event, entity eventhorizon.Entity,
) (eventhorizon.Entity, error) {
	tx, ok := entity.(*repository.Transaction)
	if !ok {
		return nil, fmt.Errorf("unknown entity type %T", entity)
	}
	switch e := event.Data().(type) {
	case *transaction.TransactionStarted:
		tx.ID = event.AggregateID()
		tx.Sender = e.Sender
		tx.SenderBank = e.SenderBank
		tx.Receiver = e.Receiver
		tx.ReceiverBank = e.ReceiverBank

		tx.CreatedAt = event.Timestamp()
		tx.Amount = e.Amount
		tx.Description = e.Description
		tx.Status = transaction.Started

	case *transaction.TransactionConfirmed:
		tx.Status = transaction.Confirmed

	case *transaction.TransactionCompleted:
		tx.Status = transaction.Completed

	case *transaction.TransactionFailed:
		tx.Status = transaction.Failed
		tx.ReasonForFailing = e.Reason

	default:
		return nil, fmt.Errorf("unknown event type %s/%T", event.EventType(), event.Data())
	}
	tx.UpdatedAt = event.Timestamp()
	return tx, nil
}
