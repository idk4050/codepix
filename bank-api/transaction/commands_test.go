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
	ID := uuid.New()
	ctx := context.Background()
	testCases := []struct {
		initialState *transaction.Aggregate
		err          error
	}{
		{transaction.New(ID), nil},
		{StartedTransaction(ID), transaction.ErrAlreadyStarted},
		{ConfirmedTransaction(ID), transaction.ErrAlreadyStarted},
		{CompletedTransaction(ID), transaction.ErrAlreadyStarted},
		{FailedTransaction(ID), transaction.ErrAlreadyStarted},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			ag := Copy(tc.initialState)
			cmd := ValidStartCommand(ID)

			err := ag.HandleCommand(ctx, cmd)
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				require.IsType(t, &aggregates.InvariantViolation{}, err)
				require.ErrorContains(t, err, tc.err.Error())
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
	ID := uuid.New()
	ctx := context.Background()
	testCases := []struct {
		initialState *transaction.Aggregate
		err          error
	}{
		{transaction.New(ID), transaction.ErrCannotConfirmIfNotStarted},
		{StartedTransaction(ID), nil},
		{ConfirmedTransaction(ID), transaction.ErrCannotConfirmIfNotStarted},
		{CompletedTransaction(ID), transaction.ErrCannotConfirmIfNotStarted},
		{FailedTransaction(ID), transaction.ErrCannotConfirmIfNotStarted},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			ag := Copy(tc.initialState)
			cmd := ValidConfirmCommand(ID)

			err := ag.HandleCommand(ctx, cmd)
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				require.IsType(t, &aggregates.InvariantViolation{}, err)
				require.ErrorContains(t, err, tc.err.Error())
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
	ID := uuid.New()
	ctx := context.Background()
	testCases := []struct {
		initialState *transaction.Aggregate
		err          error
	}{
		{transaction.New(ID), transaction.ErrCannotCompleteIfNotConfirmed},
		{StartedTransaction(ID), transaction.ErrCannotCompleteIfNotConfirmed},
		{ConfirmedTransaction(ID), nil},
		{CompletedTransaction(ID), transaction.ErrCannotCompleteIfNotConfirmed},
		{FailedTransaction(ID), transaction.ErrCannotCompleteIfNotConfirmed},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			ag := Copy(tc.initialState)
			cmd := ValidCompleteCommand(ID)

			err := ag.HandleCommand(ctx, cmd)
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				require.IsType(t, &aggregates.InvariantViolation{}, err)
				require.ErrorContains(t, err, tc.err.Error())
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
	ID := uuid.New()
	ctx := context.Background()
	testCases := []struct {
		initialState *transaction.Aggregate
		err          error
	}{
		{transaction.New(ID), transaction.ErrCannotFailIfNotStartedOrConfirmed},
		{StartedTransaction(ID), nil},
		{ConfirmedTransaction(ID), nil},
		{CompletedTransaction(ID), transaction.ErrCannotFailIfNotStartedOrConfirmed},
		{FailedTransaction(ID), transaction.ErrCannotFailIfNotStartedOrConfirmed},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			ag := Copy(tc.initialState)
			cmd := ValidFailCommand(ID)

			err := ag.HandleCommand(ctx, cmd)
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				require.IsType(t, &aggregates.InvariantViolation{}, err)
				require.ErrorContains(t, err, tc.err.Error())
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
