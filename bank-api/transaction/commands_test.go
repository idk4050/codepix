package transaction_test

import (
	"codepix/bank-api/lib/aggregates"
	"codepix/bank-api/transaction"
	"codepix/bank-api/transaction/transactiontest"
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var Copy = transactiontest.Copy
var Comparer = transactiontest.Comparer
var ExceptStatus = transactiontest.ExceptStatus
var ValidStartCommand = transactiontest.ValidStartCommand
var ValidConfirmCommand = transactiontest.ValidConfirmCommand
var ValidCompleteCommand = transactiontest.ValidCompleteCommand
var ValidFailCommand = transactiontest.ValidFailCommand
var StartedTransaction = transactiontest.StartedTransaction
var ConfirmedTransaction = transactiontest.ConfirmedTransaction
var CompletedTransaction = transactiontest.CompletedTransaction
var FailedTransaction = transactiontest.FailedTransaction

func TestStartTransaction(t *testing.T) {
	ctx := context.Background()
	ID := uuid.New()
	valid := ValidStartCommand(ID)

	notTheSender := ValidStartCommand(ID)
	notTheSender.BankID = valid.ReceiverBank

	testCases := []struct {
		initialState *transaction.Aggregate
		cmd          transaction.Start
		err          error
	}{
		{transaction.New(ID), valid, nil},
		{StartedTransaction(ID), valid, transaction.ErrAlreadyStarted},
		{ConfirmedTransaction(ID), valid, transaction.ErrAlreadyStarted},
		{CompletedTransaction(ID), valid, transaction.ErrAlreadyStarted},
		{FailedTransaction(ID), valid, transaction.ErrAlreadyStarted},

		{transaction.New(ID), notTheSender, transaction.ErrCannotStartIfNotTheSender},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			ag := Copy(tc.initialState)

			err := ag.HandleCommand(ctx, tc.cmd)
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				require.IsType(t, &aggregates.InvariantViolation{}, err)
				require.ErrorAs(t, err, &tc.err)
			}

			if tc.err == nil {
				lastEvent := ag.UncommittedEvents()[len(ag.UncommittedEvents())-1]
				ag.ApplyEvent(ctx, lastEvent)
				assert.Empty(t, cmp.Diff(StartedTransaction(ID), ag, Comparer, ExceptStatus))
			} else {
				assert.Empty(t, cmp.Diff(tc.initialState, ag, Comparer))
			}
		})
	}
}

func TestConfirmTransaction(t *testing.T) {
	ctx := context.Background()
	ID := uuid.New()
	valid := ValidConfirmCommand(ID)

	notTheReceiver := ValidConfirmCommand(ID)
	notTheReceiver.BankID = uuid.New()

	testCases := []struct {
		initialState *transaction.Aggregate
		cmd          transaction.Confirm
		err          error
	}{
		{transaction.New(ID), valid, transaction.ErrCannotConfirmIfNotStarted},
		{StartedTransaction(ID), valid, nil},
		{ConfirmedTransaction(ID), valid, transaction.ErrCannotConfirmIfNotStarted},
		{CompletedTransaction(ID), valid, transaction.ErrCannotConfirmIfNotStarted},
		{FailedTransaction(ID), valid, transaction.ErrCannotConfirmIfNotStarted},

		{StartedTransaction(ID), notTheReceiver, transaction.ErrCannotConfirmIfNotTheReceiver},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			ag := Copy(tc.initialState)

			err := ag.HandleCommand(ctx, tc.cmd)
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				require.IsType(t, &aggregates.InvariantViolation{}, err)
				require.ErrorAs(t, err, &tc.err)
			}

			if tc.err == nil {
				lastEvent := ag.UncommittedEvents()[len(ag.UncommittedEvents())-1]
				ag.ApplyEvent(ctx, lastEvent)
				assert.Empty(t, cmp.Diff(ConfirmedTransaction(ID), ag, Comparer, ExceptStatus))
			} else {
				assert.Empty(t, cmp.Diff(tc.initialState, ag, Comparer))
			}
		})
	}
}

func TestCompleteTransaction(t *testing.T) {
	ctx := context.Background()
	ID := uuid.New()
	valid := ValidCompleteCommand(ID)

	notTheSender := ValidCompleteCommand(ID)
	notTheSender.BankID = uuid.New()

	testCases := []struct {
		initialState *transaction.Aggregate
		cmd          transaction.Complete
		err          error
	}{
		{transaction.New(ID), valid, transaction.ErrCannotCompleteIfNotConfirmed},
		{StartedTransaction(ID), valid, transaction.ErrCannotCompleteIfNotConfirmed},
		{ConfirmedTransaction(ID), valid, nil},
		{CompletedTransaction(ID), valid, transaction.ErrCannotCompleteIfNotConfirmed},
		{FailedTransaction(ID), valid, transaction.ErrCannotCompleteIfNotConfirmed},

		{ConfirmedTransaction(ID), notTheSender, transaction.ErrCannotCompleteIfNotTheSender},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			ag := Copy(tc.initialState)

			err := ag.HandleCommand(ctx, tc.cmd)
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				require.IsType(t, &aggregates.InvariantViolation{}, err)
				require.ErrorAs(t, err, &tc.err)
			}

			if tc.err == nil {
				lastEvent := ag.UncommittedEvents()[len(ag.UncommittedEvents())-1]
				ag.ApplyEvent(ctx, lastEvent)
				assert.Empty(t, cmp.Diff(CompletedTransaction(ID), ag, Comparer, ExceptStatus))
			} else {
				assert.Empty(t, cmp.Diff(tc.initialState, ag, Comparer))
			}
		})
	}
}

func TestFailTransaction(t *testing.T) {
	ctx := context.Background()
	ID := uuid.New()
	valid := ValidFailCommand(ID)

	notSenderOrReceiver := ValidFailCommand(ID)
	notSenderOrReceiver.BankID = uuid.New()

	testCases := []struct {
		initialState *transaction.Aggregate
		cmd          transaction.Fail
		err          error
	}{
		{transaction.New(ID), valid, transaction.ErrCannotFailIfNotStartedOrConfirmed},
		{StartedTransaction(ID), valid, nil},
		{ConfirmedTransaction(ID), valid, nil},
		{CompletedTransaction(ID), valid, transaction.ErrCannotFailIfNotStartedOrConfirmed},
		{FailedTransaction(ID), valid, transaction.ErrCannotFailIfNotStartedOrConfirmed},

		{StartedTransaction(ID), notSenderOrReceiver, transaction.ErrCannotFailIfNotSenderOrReceiver},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			ag := Copy(tc.initialState)

			err := ag.HandleCommand(ctx, tc.cmd)
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				require.IsType(t, &aggregates.InvariantViolation{}, err)
				require.ErrorAs(t, err, &tc.err)
			}

			if tc.err == nil {
				lastEvent := ag.UncommittedEvents()[len(ag.UncommittedEvents())-1]
				ag.ApplyEvent(ctx, lastEvent)
				failedTransaction := FailedTransaction(ID)
				assert.Empty(t, cmp.Diff(failedTransaction.Transaction, ag.Transaction, ExceptStatus))
			} else {
				assert.Empty(t, cmp.Diff(tc.initialState, ag, Comparer))
			}
		})
	}
}
