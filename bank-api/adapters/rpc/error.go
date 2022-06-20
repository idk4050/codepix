package rpc

import (
	"codepix/bank-api/adapters/rpc/locale"
	"codepix/bank-api/lib/aggregates"
	"codepix/bank-api/lib/repositories"
	"codepix/bank-api/lib/validation"
	"context"
	"errors"
	"sort"

	"github.com/looplab/eventhorizon"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Mapping map[any]codes.Code

func MapError(ctx context.Context, err error, errorMappings ...Mapping) error {
	if err == nil {
		return nil
	}
	if _, alreadyStatus := status.FromError(err); alreadyStatus {
		return err
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
	case *repositories.NotFoundError:
		return status.Error(codes.NotFound, err.Error())

	case *repositories.AlreadyExistsError:
		return status.Error(codes.AlreadyExists, err.Error())

	case *repositories.InternalError:
		return status.Error(codes.Internal, err.Error())

	case *aggregates.InvariantViolation:
		switch err := err.Err.(type) {

		case *aggregates.StatusMismatchError:
			return status.Error(codes.Aborted, err.Error())

		case *aggregates.PermissionError:
			return status.Error(codes.PermissionDenied, err.Error())

		default:
			return MapError(ctx, err, errorMappings...)
		}

	case *eventhorizon.AggregateError:
		return MapError(ctx, err.Err, errorMappings...)

	default:
		return err
	}
}

func ValidationError(validator *validation.Validator,
	ctx context.Context, validationError *validation.Error,
) *status.Status {
	locales := locale.FromContext(ctx)
	errorMap := validator.MapErrors(validationError, locales...)
	errMsg := ValidationErrorMessage(errorMap)

	status := status.New(codes.InvalidArgument, validationError.Error())
	statusWithDetails, err := status.WithDetails(errMsg)
	if err == nil {
		return statusWithDetails
	}
	return status
}

func ValidationErrorMessage(errorMap map[string]string) *errdetails.BadRequest {
	fields := []string{}
	for field := range errorMap {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	violations := []*errdetails.BadRequest_FieldViolation{}
	for _, field := range fields {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       field,
			Description: errorMap[field],
		})
	}
	return &errdetails.BadRequest{
		FieldViolations: violations,
	}
}
