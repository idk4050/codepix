package eventrepositories

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ErrorAggregate struct {
	Type    string
	ID      uuid.UUID
	Version uint
}

type NotFoundError struct {
	ErrorAggregate
	When time.Time
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %s not found", e.Type, e.ID)
}
func NewNotFoundError(aggregate ErrorAggregate) *NotFoundError {
	return &NotFoundError{aggregate, time.Now()}
}

type VersionConflictError struct {
	ErrorAggregate
	When time.Time
}

func (e *VersionConflictError) Error() string {
	return fmt.Sprintf("version conflict in %s %s version %d", e.Type, e.ID, e.Version)
}
func NewVersionConflictError(aggregate ErrorAggregate) *VersionConflictError {
	return &VersionConflictError{aggregate, time.Now()}
}

type InternalError struct {
	ErrorAggregate
	Operation string
	Message   string
	When      time.Time
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("event repository error: %s on %s %s version %d: %s",
		e.Operation, e.Type, e.ID, e.Version, e.Message)
}
func NewInternalError(aggregate ErrorAggregate, operation string, message string) *InternalError {
	return &InternalError{aggregate, operation, message, time.Now()}
}
