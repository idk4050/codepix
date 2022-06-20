package service_test

import (
	"codepix/bank-api/adapters/rpc"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/bankapitest"
	"codepix/bank-api/lib/aggregates"
	"codepix/bank-api/lib/repositories"
	"codepix/bank-api/pixkey/pixkeytest"
	pixkeyrepository "codepix/bank-api/pixkey/repository"
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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
)

type SenderIDs = transactiontest.SenderIDs

var ValidPixKey = pixkeytest.ValidPixKey
var ValidStartRequest = transactiontest.ValidStartRequest
var InvalidStartRequest = transactiontest.InvalidStartRequest
var ExceptID = transactiontest.ExceptID
var Service = transactiontest.WriteService
var ServiceWithMocks = transactiontest.WriteServiceWithMocks
var AuthenticatedContext = bankapitest.AuthenticatedContext

const projectionTimeout = time.Millisecond * 100
const projectionInterval = time.Millisecond * 20

func TestStart(t *testing.T) {
	type request = proto.StartRequest
	type reply = proto.Started
	type command = transaction.Start

	type findReceiver = []any

	type in struct {
		ctx     context.Context
		request *request
		command *command
	}
	type out struct {
		findReceiver *findReceiver
		err          error
		status       *status.Status
	}
	type testCase struct {
		description string
		in          in
		out         out
	}

	client, commandHandler, pixKeyRepo := ServiceWithMocks()

	pixKey := ValidPixKey()
	receiver := &pixKey

	senderIDs := &SenderIDs{
		AccountID: uuid.New(),
		BankID:    uuid.New(),
	}
	receiverIDs := &pixkeyrepository.IDs{
		PixKeyID:  uuid.New(),
		AccountID: uuid.New(),
		BankID:    uuid.New(),
	}

	validRequest := ValidStartRequest()
	validRequest.SenderId = senderIDs.AccountID[:]
	validCommand := &command{
		BankID:       senderIDs.BankID,
		Sender:       senderIDs.AccountID,
		SenderBank:   senderIDs.BankID,
		Receiver:     receiverIDs.AccountID,
		ReceiverBank: receiverIDs.BankID,
		Amount:       validRequest.Amount,
		Description:  validRequest.Description,
	}

	ctx := AuthenticatedContext(context.Background(), validCommand.SenderBank)
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
				&findReceiver{receiver, receiverIDs, nil},
				nil,
				status.New(codes.OK, ""),
			},
		},
		{
			"invalid",
			in{
				ctxWithLocale,
				InvalidStartRequest(),
				nil,
			},
			out{
				nil,
				nil,
				func() *status.Status {
					status, _ := status.New(codes.InvalidArgument,
						"validation failed on sender_id (required)").
						WithDetails(rpc.ValidationErrorMessage(map[string]string{
							"amount":       "Amount is a required field",
							"description":  "Description must be a maximum of 100 characters in length",
							"receiver_key": "Receiver key is a required field",
							"sender_id":    "Sender ID is a required field",
						}))
					return status
				}(),
			},
		},
		{
			"receiver not found",
			in{
				ctx,
				validRequest,
				nil,
			},
			out{
				&findReceiver{nil, nil, &repositories.NotFoundError{}},
				nil,
				status.New(codes.NotFound, ""),
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
				&findReceiver{receiver, receiverIDs, nil},
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
				&findReceiver{receiver, receiverIDs, nil},
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
				nil,
				status.New(codes.Unauthenticated, ""),
			},
		},
		{
			"find receiver internal error",
			in{
				ctx,
				validRequest,
				nil,
			},
			out{
				&findReceiver{nil, nil, &repositories.InternalError{}},
				nil,
				status.New(codes.Internal, ""),
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
				&findReceiver{receiver, receiverIDs, nil},
				errors.New("some error"),
				status.New(codes.Unknown, ""),
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			if tc.out.findReceiver != nil {
				pixKeyRepo.On("FindByKey", tc.in.request.ReceiverKey).
					Return(*tc.out.findReceiver...).Once()
			}
			if tc.in.command != nil {
				commandHandler.On("HandleCommand", mock.IsType(tc.in.ctx),
					mock.MatchedBy(func(cmd command) bool { return cmp.Equal(*tc.in.command, cmd, ExceptID) })).
					Return(tc.out.err).Once()
			}

			reply, err := client.Start(tc.in.ctx, tc.in.request)

			status, _ := status.FromError(err)
			assert.Equal(t, tc.out.status.Code().String(), status.Code().String())

			if tc.out.status.Code() == codes.OK {
				require.NotNil(t, reply)
				assert.NotNil(t, reply.Id)
			}
			if tc.out.status.Message() != "" {
				assert.Empty(t, cmp.Diff(tc.out.status.Proto(), status.Proto(), protocmp.Transform()))
			}
		})
	}
}

func TestStartIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	type request = proto.StartRequest
	type reply = proto.Started

	type testCase struct {
		description  string
		request      func() (context.Context, *request, uuid.UUID)
		status       *status.Status
		persistCheck func(t *testing.T, request *request, reply *reply, bankID uuid.UUID)
	}

	client, repo, pixKeyRepo, creator, tearDown := Service()
	defer tearDown()

	now := time.Now().Truncate(time.Millisecond)

	shouldPersist := func(t *testing.T, request *request, reply *reply, bankID uuid.UUID) {
		ID := *(*uuid.UUID)(reply.Id)

		_, receiverIDs, err := pixKeyRepo.FindByKey(request.ReceiverKey)
		require.NotNil(t, receiverIDs)
		require.NoError(t, err)

		assert.Eventually(t, func() bool {
			tx, _ := repo.Find(context.Background(), ID)
			return tx != nil &&
				tx.ID == ID &&
				tx.CreatedAt.After(now) &&
				tx.UpdatedAt.After(now) &&
				tx.Sender == *(*uuid.UUID)(request.SenderId) &&
				tx.SenderBank == bankID &&
				tx.Receiver == receiverIDs.AccountID &&
				tx.ReceiverBank == receiverIDs.BankID &&
				tx.Amount == request.Amount &&
				tx.Description == request.Description &&
				tx.Status == transaction.Started &&
				tx.ReasonForFailing == ""
		}, projectionTimeout, projectionInterval)
	}
	shouldNotPersist := func(t *testing.T, request *request, reply *reply, bankID uuid.UUID) {
		opts := repository.ListOptions{}

		if request.SenderId != nil {
			opts.SenderID = *(*uuid.UUID)(request.SenderId)
		}
		_, receiverIDs, _ := pixKeyRepo.FindByKey(request.ReceiverKey)
		if receiverIDs != nil {
			opts.ReceiverID = receiverIDs.AccountID
		}

		assert.Never(t, func() bool {
			transactions, err := repo.List(context.Background(), opts)
			return len(transactions) > 0 && err == nil
		}, projectionTimeout, projectionInterval)
	}

	validRequest := func() (context.Context, *request, uuid.UUID) {
		senderIDs := SenderIDs{
			AccountID: uuid.New(),
			BankID:    uuid.New(),
		}
		receiver := ValidPixKey()
		creator.ReceiverIDs(receiver)

		valid := ValidStartRequest()
		request := &request{
			SenderId:    senderIDs.AccountID[:],
			ReceiverKey: receiver.Key,
			Amount:      valid.Amount,
			Description: valid.Description,
		}
		ctx := AuthenticatedContext(context.Background(), senderIDs.BankID)
		return ctx, request, senderIDs.BankID
	}
	invalidRequest := func() (context.Context, *request, uuid.UUID) {
		senderIDs := SenderIDs{
			AccountID: uuid.New(),
			BankID:    uuid.New(),
		}
		request := InvalidStartRequest()
		request.SenderId = senderIDs.AccountID[:]

		ctx := AuthenticatedContext(context.Background(), uuid.New())
		ctx = metadata.AppendToOutgoingContext(ctx, "locale", validator.EN_US)
		return ctx, request, uuid.New()
	}
	receiverNotFoundRequest := func() (context.Context, *request, uuid.UUID) {
		senderIDs := SenderIDs{
			AccountID: uuid.New(),
			BankID:    uuid.New(),
		}
		receiver := ValidPixKey()

		valid := ValidStartRequest()
		request := &request{
			SenderId:    senderIDs.AccountID[:],
			ReceiverKey: receiver.Key,
			Amount:      valid.Amount,
			Description: valid.Description,
		}
		ctx := AuthenticatedContext(context.Background(), senderIDs.BankID)
		return ctx, request, senderIDs.BankID
	}
	unauthenticatedRequest := func() (context.Context, *request, uuid.UUID) {
		senderIDs := SenderIDs{
			AccountID: uuid.New(),
			BankID:    uuid.New(),
		}
		request := ValidStartRequest()
		request.SenderId = senderIDs.AccountID[:]

		return context.Background(), request, uuid.Nil
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
					"validation failed on receiver_key (required)").
					WithDetails(rpc.ValidationErrorMessage(map[string]string{
						"amount":       "Amount is a required field",
						"description":  "Description must be a maximum of 100 characters in length",
						"receiver_key": "Receiver key is a required field",
					}))
				return status
			}(),
			shouldNotPersist,
		},
		{
			"receiver not found",
			receiverNotFoundRequest,
			status.New(codes.NotFound, ""),
			shouldNotPersist,
		},
		{
			"unauthenticated",
			unauthenticatedRequest,
			status.New(codes.Unauthenticated, ""),
			shouldNotPersist,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			ctx, request, bankID := tc.request()

			reply, err := client.Start(ctx, request)

			status, _ := status.FromError(err)
			assert.Equal(t, tc.status.Code().String(), status.Code().String())

			if tc.status.Code() == codes.OK {
				assert.NotNil(t, reply)
			}
			if tc.status.Message() != "" {
				assert.JSONEq(t, protojson.Format(tc.status.Proto()), protojson.Format(status.Proto()))
			}
			tc.persistCheck(t, request, reply, bankID)
		})
	}
}
