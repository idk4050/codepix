package projectionclient

import (
	"codepix/bank-api/lib/repositories"
	"errors"

	"github.com/looplab/eventhorizon"
)

func MapError(err error, entityType string) error {
	switch err := err.(type) {
	case nil:
		return nil
	case *eventhorizon.RepoError:
		switch {
		case errors.Is(err.Err, eventhorizon.ErrEntityNotFound):
			return &repositories.NotFoundError{entityType}
		default:
			return &repositories.InternalError{string(err.Op), entityType, err.Error()}
		}
	}
	return &repositories.InternalError{"unknown operation", entityType, err.Error()}
}
