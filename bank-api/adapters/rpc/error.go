package rpc

import (
	"codepix/bank-api/adapters/rpc/locale"
	"codepix/bank-api/adapters/rpc/proto"
	"codepix/bank-api/lib/aggregates"
	"codepix/bank-api/lib/eventrepositories"
	"codepix/bank-api/lib/repositories"
	"codepix/bank-api/lib/validation"
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Mapping map[any]codes.Code

func MapError(ctx context.Context, err error, errorMappings ...Mapping) error {
	if err == nil {
		return nil
	}
	for _, errorMapping := range errorMappings {
		for mappedError, mappedStatusCode := range errorMapping {
			_, ok := mappedError.(error)

			if ok && errors.As(err, &mappedError) {
				return status.Error(mappedStatusCode, mappedError.(error).Error())
			}
		}
	}
	switch err := err.(type) {
	case *aggregates.InvariantViolation:
		return status.Error(codes.InvalidArgument, err.Error())

	case *repositories.NotFoundError,
		*eventrepositories.NotFoundError:
		return status.Error(codes.NotFound, err.Error())

	case *repositories.AlreadyExistsError:
		return status.Error(codes.AlreadyExists, err.Error())

	case *eventrepositories.VersionConflictError:
		return status.Error(codes.Aborted, err.Error())

	case *repositories.InternalError,
		*eventrepositories.InternalError:
		return status.Error(codes.Internal, "internal error")

	default:
		return status.Error(codes.Unknown, "unknown error")
	}
}

func ValidationError(validator *validation.Validator,
	ctx context.Context, validationError *validation.Error,
) *status.Status {
	locales := locale.FromContext(ctx)

	errMsg := &proto.ValidationError{
		Errors: validator.MapErrors(validationError, locales...),
	}
	status := status.New(codes.InvalidArgument, validationError.Error())
	statusWithDetails, err := status.WithDetails(errMsg)
	if err == nil {
		return statusWithDetails
	}
	return status
}
