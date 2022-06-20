package projection_test

import (
	"codepix/bank-api/lib/repositories"
	"codepix/bank-api/transaction"
	"codepix/bank-api/transaction/transactiontest"
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ValidStartCommand = transactiontest.ValidStartCommand
var ValidConfirmCommand = transactiontest.ValidConfirmCommand
var ValidCompleteCommand = transactiontest.ValidCompleteCommand
var ValidFailCommand = transactiontest.ValidFailCommand
var ReadRepo = transactiontest.ReadRepo

func Test(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	readRepo, commandHandler, tearDown := ReadRepo()
	defer tearDown()

	var timeout = time.Millisecond * 100
	var interval = time.Millisecond * 20

	TestCompletedTransaction := func(t *testing.T) {
		ID := uuid.New()
		now := time.Now().Truncate(time.Millisecond)
		ctx := context.Background()

		start := ValidStartCommand(ID)
		confirm := ValidConfirmCommand(ID)
		complete := ValidCompleteCommand(ID)

		err := commandHandler.HandleCommand(ctx, start)
		require.NoError(t, err)
		err = commandHandler.HandleCommand(ctx, confirm)
		require.NoError(t, err)
		err = commandHandler.HandleCommand(ctx, complete)
		require.NoError(t, err)

		assert.Eventually(t, func() bool {
			tx, err := readRepo.Find(ID)
			return assert.NotNil(t, tx) &&
				assert.NoError(t, err) &&
				assert.Equal(t, ID, tx.ID) &&
				assert.GreaterOrEqual(t, tx.CreatedAt, now) &&
				assert.GreaterOrEqual(t, tx.UpdatedAt, now) &&
				assert.Equal(t, start.Sender, tx.Sender) &&
				assert.Equal(t, start.SenderBank, tx.SenderBank) &&
				assert.Equal(t, start.Receiver, tx.Receiver) &&
				assert.Equal(t, start.ReceiverBank, tx.ReceiverBank) &&
				assert.Equal(t, start.Amount, tx.Amount) &&
				assert.Equal(t, start.Description, tx.Description) &&
				assert.Equal(t, transaction.Completed, tx.Status) &&
				assert.Empty(t, tx.ReasonForFailing)
		}, timeout, interval)
	}

	TestFailedTransaction := func(t *testing.T) {
		ID := uuid.New()
		now := time.Now().Truncate(time.Millisecond)
		ctx := context.Background()

		start := ValidStartCommand(ID)
		confirm := ValidConfirmCommand(ID)
		fail := ValidFailCommand(ID)

		err := commandHandler.HandleCommand(ctx, start)
		require.NoError(t, err)
		err = commandHandler.HandleCommand(ctx, confirm)
		require.NoError(t, err)
		err = commandHandler.HandleCommand(ctx, fail)
		require.NoError(t, err)

		assert.Eventually(t, func() bool {
			tx, err := readRepo.Find(ID)
			return assert.NotNil(t, tx) &&
				assert.NoError(t, err) &&
				assert.Equal(t, ID, tx.ID) &&
				assert.GreaterOrEqual(t, tx.CreatedAt, now) &&
				assert.GreaterOrEqual(t, tx.UpdatedAt, now) &&
				assert.Equal(t, start.Sender, tx.Sender) &&
				assert.Equal(t, start.SenderBank, tx.SenderBank) &&
				assert.Equal(t, start.Receiver, tx.Receiver) &&
				assert.Equal(t, start.ReceiverBank, tx.ReceiverBank) &&
				assert.Equal(t, start.Amount, tx.Amount) &&
				assert.Equal(t, start.Description, tx.Description) &&
				assert.Equal(t, transaction.Failed, tx.Status) &&
				assert.Equal(t, fail.Reason, tx.ReasonForFailing)
		}, timeout, interval)
	}

	TestInvalidCommandOrder := func(t *testing.T) {
		ID := uuid.New()
		now := time.Now().Truncate(time.Millisecond)
		ctx := context.Background()

		start := ValidStartCommand(ID)
		complete := ValidCompleteCommand(ID)

		err := commandHandler.HandleCommand(ctx, start)
		require.NoError(t, err)
		err = commandHandler.HandleCommand(ctx, complete)
		require.Error(t, err)

		assert.Eventually(t, func() bool {
			tx, err := readRepo.Find(ID)
			return assert.NotNil(t, tx) &&
				assert.NoError(t, err) &&
				assert.Equal(t, ID, tx.ID) &&
				assert.GreaterOrEqual(t, tx.CreatedAt, now) &&
				assert.GreaterOrEqual(t, tx.UpdatedAt, now) &&
				assert.Equal(t, start.Sender, tx.Sender) &&
				assert.Equal(t, start.SenderBank, tx.SenderBank) &&
				assert.Equal(t, start.Receiver, tx.Receiver) &&
				assert.Equal(t, start.ReceiverBank, tx.ReceiverBank) &&
				assert.Equal(t, start.Amount, tx.Amount) &&
				assert.Equal(t, start.Description, tx.Description) &&
				assert.Equal(t, transaction.Started, tx.Status) &&
				assert.Empty(t, tx.ReasonForFailing)
		}, timeout, interval)
	}

	TestMissingTransaction := func(t *testing.T) {
		missingID := uuid.New()
		assert.Eventually(t, func() bool {
			tx, err := readRepo.Find(missingID)
			return assert.Nil(t, tx) &&
				assert.Error(t, err) &&
				assert.IsType(t, &repositories.NotFoundError{}, err)
		}, timeout, interval)
	}

	tests := []func(t *testing.T){
		TestCompletedTransaction,
		TestFailedTransaction,
		TestInvalidCommandOrder,
		TestMissingTransaction,
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), test)
	}
}
