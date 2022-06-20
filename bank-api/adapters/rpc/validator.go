package rpc

import (
	"codepix/bank-api/adapters/modifier"
	"codepix/bank-api/lib/validation"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryValidator(validator *validation.Validator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler,
	) (any, error) {
		err := validate(validator, ctx, req)
		if err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

func StreamValidator(validator *validation.Validator) grpc.StreamServerInterceptor {
	return func(server any, stream grpc.ServerStream,
		info *grpc.StreamServerInfo, next grpc.StreamHandler) error {
		return next(server, &validatedStream{stream, validator})
	}
}

type validatedStream struct {
	grpc.ServerStream
	validator *validation.Validator
}

func (s *validatedStream) RecvMsg(msg interface{}) error {
	err := s.ServerStream.RecvMsg(msg)
	if err != nil {
		return err
	}
	return validate(s.validator, s.Context(), msg)
}

func validate(validator *validation.Validator, ctx context.Context, req any) error {
	err := modifier.Mold(req)
	if err != nil {
		return status.Errorf(codes.Internal, "validate request: %s", err.Error())
	}
	err = validation.Validate(validator, req)
	if err != nil {
		switch err := err.(type) {
		case *validation.Error:
			return ValidationError(validator, ctx, err).Err()
		default:
			status.Errorf(codes.Internal, "validate request: %s", err.Error())
		}
	}
	return nil
}
