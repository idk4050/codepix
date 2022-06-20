package projection_test

import (
	"codepix/bank-api/lib/repositories"
	"codepix/bank-api/transaction"
	"codepix/bank-api/transaction/read/repository"
	"codepix/bank-api/transaction/transactiontest"
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ValidStartCommand = transactiontest.ValidStartCommand
var ValidConfirmCommand = transactiontest.ValidConfirmCommand
var ValidCompleteCommand = transactiontest.ValidCompleteCommand
var ValidFailCommand = transactiontest.ValidFailCommand
var ReadRepo = transactiontest.ReadRepo

const projectionTimeout = time.Millisecond * 150
const projectionInterval = time.Millisecond * 50

func TestProjection(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	repo, commandHandler, tearDown := ReadRepo()
	defer tearDown()

	type test struct {
		description string
		fn          func(*testing.T)
	}
	tests := []test{
		{"find", Find(repo, commandHandler)},
		{"list", List(repo, commandHandler)},
	}
	for i, test := range tests {
		t.Run(fmt.Sprint(i, "_", test.description), test.fn)
	}
}

func Find(repo repository.Repository, commandHandler eventhorizon.CommandHandler) func(t *testing.T) {
	return func(t *testing.T) {
		if testing.Short() {
			t.Skip()
		}

		TestCompletedTransaction := func(t *testing.T) {
			ctx := context.Background()
			ID := uuid.New()
			now := time.Now().Truncate(time.Millisecond)

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
				tx, err := repo.Find(ctx, ID)
				return tx != nil &&
					err == nil &&
					tx.ID == ID &&
					!tx.CreatedAt.Before(now) &&
					!tx.UpdatedAt.Before(now) &&
					tx.Sender == start.Sender &&
					tx.SenderBank == start.SenderBank &&
					tx.Receiver == start.Receiver &&
					tx.ReceiverBank == start.ReceiverBank &&
					tx.Amount == start.Amount &&
					tx.Description == start.Description &&
					tx.Status == transaction.Completed &&
					tx.ReasonForFailing == ""
			}, projectionTimeout, projectionInterval)
		}
		TestFailedTransaction := func(t *testing.T) {
			ctx := context.Background()
			ID := uuid.New()
			now := time.Now().Truncate(time.Millisecond)

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
				tx, err := repo.Find(ctx, ID)
				return tx != nil &&
					err == nil &&
					tx.ID == ID &&
					!tx.CreatedAt.Before(now) &&
					!tx.UpdatedAt.Before(now) &&
					tx.Sender == start.Sender &&
					tx.SenderBank == start.SenderBank &&
					tx.Receiver == start.Receiver &&
					tx.ReceiverBank == start.ReceiverBank &&
					tx.Amount == start.Amount &&
					tx.Description == start.Description &&
					tx.Status == transaction.Failed &&
					tx.ReasonForFailing == fail.Reason
			}, projectionTimeout, projectionInterval)
		}
		TestInvalidCommandOrder := func(t *testing.T) {
			ctx := context.Background()
			ID := uuid.New()
			now := time.Now().Truncate(time.Millisecond)

			start := ValidStartCommand(ID)
			complete := ValidCompleteCommand(ID)

			err := commandHandler.HandleCommand(ctx, start)
			require.NoError(t, err)
			err = commandHandler.HandleCommand(ctx, complete)
			require.Error(t, err)

			assert.Eventually(t, func() bool {
				tx, err := repo.Find(ctx, ID)
				return tx != nil &&
					err == nil &&
					tx.ID == ID &&
					!tx.CreatedAt.Before(now) &&
					!tx.UpdatedAt.Before(now) &&
					tx.Sender == start.Sender &&
					tx.SenderBank == start.SenderBank &&
					tx.Receiver == start.Receiver &&
					tx.ReceiverBank == start.ReceiverBank &&
					tx.Amount == start.Amount &&
					tx.Description == start.Description &&
					tx.Status == transaction.Started &&
					tx.ReasonForFailing == ""
			}, projectionTimeout, projectionInterval)
		}
		TestMissingTransaction := func(t *testing.T) {
			ctx := context.Background()
			missingID := uuid.New()

			assert.Eventually(t, func() bool {
				tx, err := repo.Find(ctx, missingID)
				notFound := &repositories.NotFoundError{}
				return tx == nil && errors.As(err, &notFound)
			}, projectionTimeout, projectionInterval)
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
}

func List(repo repository.Repository, commandHandler eventhorizon.CommandHandler) func(t *testing.T) {
	return func(t *testing.T) {
		if testing.Short() {
			t.Skip()
		}

		TestCreatedAfterFilter := func(t *testing.T) {
			ctx := context.Background()
			totalTxs := 10
			expectedTxs := totalTxs / 2

			senderID := uuid.New()
			for i := 0; i < expectedTxs; i++ {
				start := ValidStartCommand(uuid.New())
				start.Sender = senderID
				err := commandHandler.HandleCommand(ctx, start)
				require.NoError(t, err)
			}
			now := time.Now()
			for i := 0; i < expectedTxs; i++ {
				start := ValidStartCommand(uuid.New())
				start.Sender = senderID
				err := commandHandler.HandleCommand(ctx, start)
				require.NoError(t, err)
			}

			assert.Eventually(t, func() bool {
				opts := repository.ListOptions{
					CreatedAfter: now,
					SenderID:     senderID,
				}
				txs, err := repo.List(ctx, opts)
				if len(txs) != expectedTxs || err != nil {
					return false
				}
				for _, tx := range txs {
					if tx.CreatedAt.UnixMilli() < now.UnixMilli() {
						return false
					}
				}
				return true
			}, projectionTimeout, projectionInterval)
		}
		TestIDFilters := func(t *testing.T) {
			ctx := context.Background()
			totalTxs := 10
			expectedTxs := totalTxs / 2

			senderID := uuid.New()
			for i := 0; i < expectedTxs; i++ {
				start := ValidStartCommand(uuid.New())
				start.Sender = senderID
				err := commandHandler.HandleCommand(ctx, start)
				require.NoError(t, err)
			}

			receiverID := uuid.New()
			for i := 0; i < expectedTxs; i++ {
				start := ValidStartCommand(uuid.New())
				start.Receiver = receiverID
				err := commandHandler.HandleCommand(ctx, start)
				require.NoError(t, err)
			}

			assert.Eventually(t, func() bool {
				opts := repository.ListOptions{
					SenderID: senderID,
				}
				txs, err := repo.List(ctx, opts)
				if len(txs) != expectedTxs || err != nil {
					return false
				}
				for _, tx := range txs {
					if tx.Sender != senderID {
						return false
					}
				}
				return true
			}, projectionTimeout, projectionInterval)

			assert.Eventually(t, func() bool {
				opts := repository.ListOptions{
					ReceiverID: receiverID,
				}
				txs, err := repo.List(ctx, opts)
				if len(txs) != expectedTxs || err != nil {
					return false
				}
				for _, tx := range txs {
					if tx.Receiver != receiverID {
						return false
					}
				}
				return true
			}, projectionTimeout, projectionInterval)
		}
		TestNewestFirstSort := func(t *testing.T) {
			ctx := context.Background()
			totalTxs := 10

			senderID := uuid.New()
			for i := 0; i < totalTxs; i++ {
				start := ValidStartCommand(uuid.New())
				start.Sender = senderID
				err := commandHandler.HandleCommand(ctx, start)
				require.NoError(t, err)
			}

			assert.Eventually(t, func() bool {
				txs, err := repo.List(ctx, repository.ListOptions{
					SenderID: senderID,
				})
				if len(txs) != totalTxs || err != nil {
					return false
				}
				for i := 0; i < len(txs)-1; i++ {
					cur := txs[i]
					next := txs[i+1]
					if cur.CreatedAt.Before(next.CreatedAt) {
						return false
					}
				}
				return true
			}, projectionTimeout, projectionInterval)
		}
		TestLimit := func(t *testing.T) {
			ctx := context.Background()
			totalTxs := 10
			expectedTxs := totalTxs / 2

			senderID := uuid.New()
			for i := 0; i < totalTxs; i++ {
				start := ValidStartCommand(uuid.New())
				start.Sender = senderID
				err := commandHandler.HandleCommand(ctx, start)
				require.NoError(t, err)
			}

			assert.Eventually(t, func() bool {
				txs, err := repo.List(ctx, repository.ListOptions{
					Limit:    uint64(expectedTxs),
					SenderID: senderID,
				})
				if len(txs) != expectedTxs || err != nil {
					return false
				}
				return true
			}, projectionTimeout, projectionInterval)
		}
		TestSkip := func(t *testing.T) {
			ctx := context.Background()
			totalTxs := 10
			skip := totalTxs / 2
			expectedTxs := totalTxs - skip

			senderID := uuid.New()
			for i := 0; i < totalTxs; i++ {
				start := ValidStartCommand(uuid.New())
				start.Sender = senderID
				err := commandHandler.HandleCommand(ctx, start)
				require.NoError(t, err)
			}

			assert.Eventually(t, func() bool {
				txs, err := repo.List(ctx, repository.ListOptions{
					Skip:     uint64(skip),
					SenderID: senderID,
				})
				if len(txs) != expectedTxs || err != nil {
					return false
				}
				return true
			}, projectionTimeout, projectionInterval)
		}
		tests := []func(t *testing.T){
			TestCreatedAfterFilter,
			TestIDFilters,
			TestNewestFirstSort,
			TestLimit,
			TestSkip,
		}
		for i, test := range tests {
			t.Run(strconv.Itoa(i), test)
		}
	}
}
