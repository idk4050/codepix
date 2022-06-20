package repository

import (
	"codepix/bank-api/transaction"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
)

var EntityType = "transaction"
var RepositoryType = "transactions"

type Repository interface {
	Find(ctx context.Context, ID uuid.UUID) (*Transaction, error)
	List(ctx context.Context, options ListOptions) ([]ListItem, error)
}

type Transaction struct {
	ID           uuid.UUID `bson:"_id"`
	Sender       uuid.UUID `bson:"sender"`
	SenderBank   uuid.UUID `bson:"sender_bank"`
	Receiver     uuid.UUID `bson:"receiver"`
	ReceiverBank uuid.UUID `bson:"receiver_bank"`

	CreatedAt        time.Time          `bson:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at"`
	Amount           transaction.Amount `bson:"amount"`
	Description      string             `bson:"description"`
	Status           transaction.Status `bson:"status"`
	ReasonForFailing string             `bson:"reason_for_failing"`
}

var _ eventhorizon.Entity = Transaction{}

func (t Transaction) EntityID() uuid.UUID {
	return t.ID
}

type ListItem = Transaction

type ListOptions struct {
	CreatedAfter time.Time
	SenderID     uuid.UUID
	ReceiverID   uuid.UUID
	Limit        uint64
	Skip         uint64
}
