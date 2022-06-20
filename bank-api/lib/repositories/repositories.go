package repositories

import (
	"fmt"
	"time"
)

type NotFoundError struct {
	What string
	When time.Time
}

func (e *NotFoundError) Error() string {
	return e.What + " not found"
}
func NewNotFoundError(what string) *NotFoundError {
	return &NotFoundError{what, time.Now()}
}

type AlreadyExistsError struct {
	What string
	When time.Time
}

func (e *AlreadyExistsError) Error() string {
	return e.What + " already exists"
}
func NewAlreadyExistsError(what string) *AlreadyExistsError {
	return &AlreadyExistsError{what, time.Now()}
}

type InternalError struct {
	Operation  string
	Collection string
	Message    string
	When       time.Time
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("repository error: %s on %s: %s", e.Operation, e.Collection, e.Message)
}
func NewInternalError(operation, collection, message string) *InternalError {
	return &InternalError{operation, collection, message, time.Now()}
}
