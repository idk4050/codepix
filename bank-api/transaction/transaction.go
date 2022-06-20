package transaction

import (
	"github.com/google/uuid"
)

type Status uint8
type Amount = uint64

const (
	Started Status = iota + 1
	Confirmed
	Completed
	Failed
)

type Transaction struct {
	Sender       uuid.UUID
	SenderBank   uuid.UUID
	Receiver     uuid.UUID
	ReceiverBank uuid.UUID
	Amount       Amount
	Description  string
	Status       Status
}
