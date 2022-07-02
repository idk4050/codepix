package aggregates

type InvariantViolation struct {
	Err error
}

func (e *InvariantViolation) Error() string {
	return "invariant violation: " + e.Err.Error()
}

type StatusMismatchError struct {
	Message string
}

func (e *StatusMismatchError) Error() string {
	return "status mismatch: " + e.Message
}

type PermissionError struct {
	Message string
}

func (e *PermissionError) Error() string {
	return "permission error: " + e.Message
}
