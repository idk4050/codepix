package stream_test

import (
	"codepix/bank-api/adapters/rpc"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/lib/aggregates"
	proto "codepix/bank-api/proto/codepix/transaction/write"
	"codepix/bank-api/transaction"
	"codepix/bank-api/transaction/read/repository"
	"codepix/bank-api/transaction/transactiontest"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

var InvalidCompleteRequest = transactiontest.InvalidCompleteRequest

func Complete(client proto.StreamClient, commandHandler *transactiontest.MockCommandHandler,
) func(t *testing.T) {
	return func(t *testing.T) {
		type request = proto.CompleteRequest
		type reply = proto.CompleteReply
		type command = transaction.Complete

		type in struct {
			ctx     context.Context
			request *request
			command *command
		}
		type out struct {
			err    error
			status *status.Status
		}
		type testCase struct {
			description string
			in          in
			out         out
		}

		ID := uuid.New()
		bankID := uuid.New()
		validRequest := &request{
			Id: ID[:],
		}
		validCommand := &command{
			ID:     ID,
			BankID: bankID,
		}

		ctx := AuthenticatedContext(context.Background(), bankID)
		ctxWithLocale := metadata.AppendToOutgoingContext(ctx, "locale", validator.EN_US)

		testCases := []testCase{
			{
				"valid",
				in{
					ctx,
					validRequest,
					validCommand,
				},
				out{
					nil,
					status.New(codes.OK, ""),
				},
			},
			{
				"invalid",
				in{
					ctxWithLocale,
					InvalidCompleteRequest(),
					nil,
				},
				out{
					nil,
					func() *status.Status {
						status, _ := status.New(codes.InvalidArgument,
							"validation failed on id (required)").
							WithDetails(rpc.ValidationErrorMessage(map[string]string{
								"id": "id is a required field",
							}))
						return status
					}(),
				},
			},
			{
				"aggregate invariant violation: status mismatch",
				in{
					ctx,
					validRequest,
					validCommand,
				},
				out{
					&aggregates.InvariantViolation{&aggregates.StatusMismatchError{}},
					status.New(codes.Aborted, ""),
				},
			},
			{
				"aggregate invariant violation: permission error",
				in{
					ctx,
					validRequest,
					validCommand,
				},
				out{
					&aggregates.InvariantViolation{&aggregates.PermissionError{}},
					status.New(codes.PermissionDenied, ""),
				},
			},
			{
				"unauthenticated",
				in{
					context.Background(),
					validRequest,
					nil,
				},
				out{
					nil,
					status.New(codes.Unauthenticated, ""),
				},
			},
			{
				"handle command internal error",
				in{
					ctx,
					validRequest,
					validCommand,
				},
				out{
					errors.New("some error"),
					status.New(codes.Unknown, ""),
				},
			},
		}
		for i, tc := range testCases {
			t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
				if tc.in.command != nil {
					commandHandler.On("HandleCommand", mock.IsType(tc.in.ctx),
						mock.MatchedBy(func(cmd command) bool { return cmp.Equal(*tc.in.command, cmd, ExceptID) })).
						Return(tc.out.err).Once()
				}

				stream, err := client.Complete(tc.in.ctx)
				require.NoError(t, err)
				err = stream.Send(tc.in.request)
				require.NoError(t, err)

				reply, err := stream.Recv()
				if err != nil {
					status, _ := status.FromError(err)
					assert.Equal(t, tc.out.status.Code().String(), status.Code().String())
				} else {
					if tc.out.status.Code() == codes.OK {
						require.NotNil(t, reply)
						require.NotNil(t, reply.GetCompleted())
					} else {
						status := reply.GetError()
						assert.Equal(t, tc.out.status.Code().String(), codes.Code(status.Code).String())
					}
				}
				if tc.out.status.Message() != "" {
					status := reply.GetError()
					assert.Empty(t, cmp.Diff(tc.out.status.Proto(), status, protocmp.Transform()))
				}
			})
		}
	}
}

func CompleteIntegration(client proto.StreamClient, repo repository.Repository,
	creator transactiontest.Creator) func(t *testing.T) {
	return func(t *testing.T) {
		if testing.Short() {
			t.Skip()
		}
		type request = proto.CompleteRequest
		type reply = proto.CompleteReply

		type testCase struct {
			description  string
			request      func() (context.Context, *request)
			status       *status.Status
			persistCheck func(t *testing.T, request *request, reply *reply)
		}

		now := time.Now().Truncate(time.Millisecond)

		shouldPersist := func(t *testing.T, request *request, reply *reply) {
			ID := *(*uuid.UUID)(request.Id)
			assert.Eventually(t, func() bool {
				tx, _ := repo.Find(context.Background(), ID)
				return tx != nil &&
					!tx.UpdatedAt.Before(now) &&
					tx.Status == transaction.Completed
			}, projectionTimeout, projectionInterval)
		}
		shouldNotPersist := func(t *testing.T, request *request, reply *reply) {
			ID, _ := uuid.FromBytes(request.Id)
			assert.Never(t, func() bool {
				tx, _ := repo.Find(context.Background(), ID)
				return tx != nil &&
					!tx.UpdatedAt.After(now) &&
					tx.Status == transaction.Started &&
					tx.ReasonForFailing == ""
			}, projectionTimeout, projectionInterval)
		}

		validRequest := func() (context.Context, *request) {
			ID, senderIDs, _ := creator.ConfirmedIDs()
			request := &request{
				Id: ID[:],
			}
			ctx := AuthenticatedContext(context.Background(), senderIDs.BankID)
			return ctx, request
		}
		invalidRequest := func() (context.Context, *request) {
			request := InvalidCompleteRequest()
			ctx := AuthenticatedContext(context.Background(), uuid.New())
			ctx = metadata.AppendToOutgoingContext(ctx, "locale", validator.EN_US)
			return ctx, request
		}
		unauthenticatedRequest := func() (context.Context, *request) {
			ID := uuid.New()
			request := &request{
				Id: ID[:],
			}
			return context.Background(), request
		}
		permissionDeniedRequest := func() (context.Context, *request) {
			ID, _, receiverIDs := creator.ConfirmedIDs()
			request := &request{
				Id: ID[:],
			}
			ctx := AuthenticatedContext(context.Background(), receiverIDs.BankID)
			return ctx, request
		}

		testCases := []testCase{
			{
				"valid",
				validRequest,
				status.New(codes.OK, ""),
				shouldPersist,
			},
			{
				"invalid",
				invalidRequest,
				func() *status.Status {
					status, _ := status.New(codes.InvalidArgument,
						"validation failed on id (required)").
						WithDetails(rpc.ValidationErrorMessage(map[string]string{
							"id": "id is a required field",
						}))
					return status
				}(),
				shouldNotPersist,
			},
			{
				"unauthenticated",
				unauthenticatedRequest,
				status.New(codes.Unauthenticated, ""),
				shouldNotPersist,
			},
			{
				"permission denied",
				permissionDeniedRequest,
				status.New(codes.PermissionDenied, ""),
				shouldNotPersist,
			},
		}
		for i, tc := range testCases {
			t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
				ctx, request := tc.request()

				stream, err := client.Complete(ctx)
				require.NoError(t, err)
				err = stream.Send(request)
				require.NoError(t, err)

				reply, err := stream.Recv()
				if err != nil {
					status, _ := status.FromError(err)
					assert.Equal(t, tc.status.Code().String(), status.Code().String())
				} else {
					if tc.status.Code() == codes.OK {
						require.NotNil(t, reply)
						require.NotNil(t, reply.GetCompleted())
					} else {
						status := reply.GetError()
						assert.Equal(t, tc.status.Code().String(), codes.Code(status.Code).String())
					}
				}
				if tc.status.Message() != "" {
					status := reply.GetError()
					assert.Empty(t, cmp.Diff(tc.status.Proto(), status, protocmp.Transform()))
				}
				tc.persistCheck(t, request, reply)
			})
		}
	}
}
