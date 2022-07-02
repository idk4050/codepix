package aggregates

import (
	"time"
)

type InvariantViolation struct {
	Message string
	When    time.Time
}

func (e *InvariantViolation) Error() string {
	return "invariant violation: " + e.Message
}
func NewInvariantViolation(message string) *InvariantViolation {
	return &InvariantViolation{message, time.Now()}
}
