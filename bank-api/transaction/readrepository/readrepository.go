package readrepository

import (
	"codepix/bank-api/transaction"
	"time"

	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
)

var EntityType = "transaction"
var RepositoryType = "transactions"

type ReadRepository interface {
	Find(ID uuid.UUID) (*Transaction, error)
}

type Transaction struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time

	Sender       uuid.UUID
	SenderBank   uuid.UUID
	Receiver     uuid.UUID
	ReceiverBank uuid.UUID
	Amount       transaction.Amount
	Description  string
	Status       transaction.Status

	ReasonForFailing string
}

var _ eventhorizon.Entity = Transaction{}

func (t Transaction) EntityID() uuid.UUID {
	return t.ID
}
