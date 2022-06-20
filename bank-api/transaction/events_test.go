package transaction_test

import (
	"codepix/bank-api/transaction"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTransactionStarted(t *testing.T) {
	tx := transaction.Transaction{}

	event := transaction.TransactionStarted{
		Sender:       uuid.New(),
		SenderBank:   uuid.New(),
		Receiver:     uuid.New(),
		ReceiverBank: uuid.New(),
		Amount:       100,
		Description:  "test",
	}
	event.Apply(&tx)

	assert.Equal(t, tx.Sender, event.Sender)
	assert.Equal(t, tx.SenderBank, event.SenderBank)
	assert.Equal(t, tx.Receiver, event.Receiver)
	assert.Equal(t, tx.ReceiverBank, event.ReceiverBank)
	assert.Equal(t, tx.Amount, event.Amount)
	assert.Equal(t, tx.Description, event.Description)
	assert.Equal(t, transaction.Started, tx.Status)
}

func TestTransactionConfirmed(t *testing.T) {
	ag := transaction.New(uuid.New())
	copy := Copy(ag)
	tx := ag.Transaction
	original := copy.Transaction

	event := transaction.TransactionConfirmed{
		SenderBank:   tx.SenderBank,
		ReceiverBank: tx.ReceiverBank,
	}
	event.Apply(tx)

	assert.Equal(t, tx.Sender, original.Sender)
	assert.Equal(t, tx.SenderBank, original.SenderBank)
	assert.Equal(t, tx.Receiver, original.Receiver)
	assert.Equal(t, tx.ReceiverBank, original.ReceiverBank)
	assert.Equal(t, tx.Amount, original.Amount)
	assert.Equal(t, tx.Description, original.Description)
	assert.Equal(t, transaction.Confirmed, tx.Status)
}

func TestTransactionCompleted(t *testing.T) {
	ag := transaction.New(uuid.New())
	copy := Copy(ag)
	tx := ag.Transaction
	original := copy.Transaction

	event := transaction.TransactionCompleted{
		SenderBank:   tx.SenderBank,
		ReceiverBank: tx.ReceiverBank,
	}
	event.Apply(tx)

	assert.Equal(t, tx.Sender, original.Sender)
	assert.Equal(t, tx.SenderBank, original.SenderBank)
	assert.Equal(t, tx.Receiver, original.Receiver)
	assert.Equal(t, tx.ReceiverBank, original.ReceiverBank)
	assert.Equal(t, tx.Amount, original.Amount)
	assert.Equal(t, tx.Description, original.Description)
	assert.Equal(t, transaction.Completed, tx.Status)
}

func TestTransactionFailed(t *testing.T) {
	ag := transaction.New(uuid.New())
	copy := Copy(ag)
	tx := ag.Transaction
	original := copy.Transaction

	event := transaction.TransactionFailed{
		SenderBank:   tx.SenderBank,
		ReceiverBank: tx.ReceiverBank,
		Reason:       "not enough balance",
	}
	event.Apply(tx)

	assert.Equal(t, tx.Sender, original.Sender)
	assert.Equal(t, tx.SenderBank, original.SenderBank)
	assert.Equal(t, tx.Receiver, original.Receiver)
	assert.Equal(t, tx.ReceiverBank, original.ReceiverBank)
	assert.Equal(t, tx.Amount, original.Amount)
	assert.Equal(t, tx.Description, original.Description)
	assert.Equal(t, transaction.Failed, tx.Status)
}
