package transaction

import "github.com/google/uuid"

const startedStream = string(StartedEvent) + "_"
const confirmedStream = string(ConfirmedEvent) + "_"
const completedStream = string(CompletedEvent) + "_"
const failedStream = string(FailedEvent) + "_"

func StartedStream(bankID uuid.UUID) string   { return startedStream + bankID.String() }
func ConfirmedStream(bankID uuid.UUID) string { return confirmedStream + bankID.String() }
func CompletedStream(bankID uuid.UUID) string { return completedStream + bankID.String() }
func FailedStream(bankID uuid.UUID) string    { return failedStream + bankID.String() }
