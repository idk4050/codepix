package repositories

import (
	"fmt"
)

type NotFoundError struct {
	What string
}

func (e NotFoundError) Error() string {
	return e.What + " not found"
}

type AlreadyExistsError struct {
	What string
}

func (e AlreadyExistsError) Error() string {
	return e.What + " already exists"
}

type InternalError struct {
	Operation  string
	Collection string
	Message    string
}

func (e InternalError) Error() string {
	return fmt.Sprintf("repository error: %s on %s: %s", e.Operation, e.Collection, e.Message)
}
