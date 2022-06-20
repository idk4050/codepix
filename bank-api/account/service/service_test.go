package service_test

import (
	"codepix/bank-api/account"
	"codepix/bank-api/account/accounttest"
	"codepix/bank-api/account/interactor"
	accountrepository "codepix/bank-api/account/repository"
	"codepix/bank-api/account/service/proto"
	rpcproto "codepix/bank-api/adapters/rpc/proto"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/bankapitest"
	"codepix/bank-api/lib/repositories"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
)

var ValidAccount = accounttest.ValidAccount
var InvalidAccount = accounttest.InvalidAccount
var Service = accounttest.Service
var ServiceWithMocks = accounttest.ServiceWithMocks
var AuthenticatedContext = bankapitest.AuthenticatedContext

func TestRegister(t *testing.T) {
	type request = proto.RegisterRequest
	type reply = proto.RegisterReply
	type input = interactor.RegisterInput
	type output = interactor.RegisterOutput

	type in struct {
		ctx     context.Context
		request *request
		input   input
	}
	type out struct {
		output *output
		err    error
		reply  *reply
		status *status.Status
	}
	type testCase struct {
		description string
		in          in
		out         out
	}

	client, interactor, _ := ServiceWithMocks()

	valid := ValidAccount()
	invalid := InvalidAccount()

	ID := uuid.New()
	bankID := uuid.New()

	ctx := AuthenticatedContext(context.Background(), bankID)
	ctxWithLocale := metadata.AppendToOutgoingContext(ctx, "locale", validator.EN_US)

	validRequest := &request{Number: valid.Number, OwnerName: valid.OwnerName}
	validInput := input{Number: valid.Number, OwnerName: valid.OwnerName, BankID: bankID}

	testCases := []testCase{
		{
			"valid",
			in{
				ctx,
				validRequest,
				validInput,
			},
			out{
				&output{Account: valid, ID: ID},
				nil,
				&reply{Id: ID[:]},
				status.New(codes.OK, ""),
			},
		},
		{
			"invalid",
			in{
				ctxWithLocale,
				&request{Number: invalid.Number, OwnerName: invalid.OwnerName},
				input{},
			},
			out{
				nil,
				nil,
				nil,
				func() *status.Status {
					status, _ := status.New(codes.InvalidArgument,
						"validation failed on number (required)").
						WithDetails(&rpcproto.ValidationError{Errors: map[string]string{
							"number":     "Account number is a required field",
							"owner_name": "Owner name is a required field",
						}})
					return status
				}(),
			},
		},
		{
			"already exists",
			in{
				ctx,
				validRequest,
				validInput,
			},
			out{
				nil,
				repositories.NewAlreadyExistsError("account"),
				nil,
				status.New(codes.AlreadyExists, ""),
			},
		},
		{
			"internal error",
			in{
				ctx,
				validRequest,
				validInput,
			},
			out{
				nil,
				repositories.NewInternalError("insert", "account", "error message"),
				nil,
				status.New(codes.Internal, ""),
			},
		},
		{
			"unknown error",
			in{
				ctx,
				validRequest,
				validInput,
			},
			out{
				nil,
				errors.New("error message"),
				nil,
				status.New(codes.Unknown, ""),
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			interactor.On("Register", tc.in.input).Return(tc.out.output, tc.out.err).Once()

			reply, err := client.Register(tc.in.ctx, tc.in.request)

			status, _ := status.FromError(err)
			assert.Equal(t, tc.out.status.Code(), status.Code())

			if tc.out.status.Code() == codes.OK {
				assert.Empty(t, cmp.Diff(tc.out.reply, reply, protocmp.Transform()))
			}
			if tc.out.status.Message() != "" {
				assert.Empty(t, cmp.Diff(tc.out.status.Proto(), status.Proto(), protocmp.Transform()))
			}
		})
	}
}

func TestRegisterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	type request = proto.RegisterRequest
	type reply = proto.RegisterReply

	type testCase struct {
		description  string
		request      func() (*request, context.Context)
		status       *status.Status
		persistCheck func(t *testing.T, request *request, reply *reply)
	}

	client, repo := Service()

	bankID := uuid.New()

	ctx := AuthenticatedContext(context.Background(), bankID)
	ctxWithLocale := metadata.AppendToOutgoingContext(ctx, "locale", validator.EN_US)

	validRequest := func() (*request, context.Context) {
		account := ValidAccount()

		request := &request{Number: account.Number, OwnerName: account.OwnerName}
		return request, ctx
	}
	invalidRequest := func() (*request, context.Context) {
		account := InvalidAccount()

		request := &request{Number: account.Number, OwnerName: account.OwnerName}
		return request, ctxWithLocale
	}
	alreadyExistsRequest := func() (*request, context.Context) {
		account := ValidAccount()
		repo.Add(account, bankID)

		request := &request{Number: account.Number, OwnerName: account.OwnerName}
		return request, ctx
	}

	noPersistCheck := func(t *testing.T, request *request, reply *reply) {}
	shouldPersist := func(t *testing.T, request *request, reply *reply) {
		ID, _ := uuid.FromBytes(reply.Id)
		_, persistedIDs, err := repo.Find(ID)
		assert.NoError(t, err)
		assert.Equal(t, ID, persistedIDs.AccountID)
	}
	shouldNotPersist := func(t *testing.T, request *request, reply *reply) {
		account := account.Account{
			Number:    request.Number,
			OwnerName: request.OwnerName,
		}
		_, err := repo.Add(account, bankID)
		assert.NoError(t, err)
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
					"validation failed on number (required)").
					WithDetails(&rpcproto.ValidationError{Errors: map[string]string{
						"number":     "Account number is a required field",
						"owner_name": "Owner name is a required field",
					}})
				return status
			}(),
			shouldNotPersist,
		},
		{
			"already exists",
			alreadyExistsRequest,
			status.New(codes.AlreadyExists, ""),
			noPersistCheck,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			request, ctx := tc.request()

			reply, err := client.Register(ctx, request)

			status, _ := status.FromError(err)
			assert.Equal(t, tc.status.Code(), status.Code())

			if tc.status.Code() == codes.OK {
				assert.NotNil(t, reply)
			}
			if tc.status.Message() != "" {
				assert.JSONEq(t, protojson.Format(tc.status.Proto()), protojson.Format(status.Proto()))
			}
			tc.persistCheck(t, request, reply)
		})
	}
}

func TestFind(t *testing.T) {
	type request = proto.FindRequest
	type reply = proto.FindReply
	type output = account.Account

	type in struct {
		ctx     context.Context
		request *request
	}
	type out struct {
		output    *output
		outputIDs *accountrepository.IDs
		err       error
		reply     *reply
		status    codes.Code
	}
	type testCase struct {
		description string
		in          in
		out         out
	}

	client, _, repo := ServiceWithMocks()

	ID := uuid.New()
	bankID := uuid.New()

	ctx := AuthenticatedContext(context.Background(), bankID)

	valid := ValidAccount()
	validRequest := &request{Id: ID[:]}

	testCases := []testCase{
		{
			"valid",
			in{
				ctx,
				validRequest,
			},
			out{
				&valid,
				&accountrepository.IDs{AccountID: ID, BankID: bankID},
				nil,
				&proto.FindReply{
					Number:    valid.Number,
					OwnerName: valid.OwnerName,
				},
				codes.OK,
			},
		},
		{
			"not found",
			in{
				ctx,
				validRequest,
			},
			out{
				nil,
				nil,
				&repositories.NotFoundError{},
				nil,
				codes.NotFound,
			},
		},
		{
			"permission denied",
			in{
				ctx,
				validRequest,
			},
			out{
				&valid,
				&accountrepository.IDs{AccountID: ID, BankID: uuid.New()},
				nil,
				nil,
				codes.PermissionDenied,
			},
		},
		{
			"unauthenticated",
			in{
				context.Background(),
				validRequest,
			},
			out{
				nil,
				nil,
				nil,
				nil,
				codes.Unauthenticated,
			},
		},
		{
			"internal error",
			in{
				ctx,
				validRequest,
			},
			out{
				nil,
				nil,
				&repositories.InternalError{},
				nil,
				codes.Internal,
			},
		},
		{
			"unknown error",
			in{
				ctx,
				validRequest,
			},
			out{
				nil,
				nil,
				errors.New("error message"),
				nil,
				codes.Unknown,
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			if tc.out.status != codes.Unauthenticated {
				repo.On("Find", ID).Return(tc.out.output, tc.out.outputIDs, tc.out.err).Once()
			}

			reply, err := client.Find(tc.in.ctx, tc.in.request)

			if tc.out.status == codes.OK {
				assert.Empty(t, cmp.Diff(tc.out.reply, reply, protocmp.Transform()))
			}
			status, _ := status.FromError(err)
			assert.Equal(t, tc.out.status, status.Code())
		})
	}
}

func TestFindIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	type request = proto.FindRequest
	type reply = proto.FindReply

	type testCase struct {
		description  string
		request      func() (context.Context, *request)
		status       codes.Code
		persistCheck func(t *testing.T, request *request, reply *reply)
	}

	client, repo := Service()

	noPersistCheck := func(t *testing.T, request *request, reply *reply) {}
	shouldExist := func(t *testing.T, request *request, reply *reply) {
		ID, _ := uuid.FromBytes(request.Id)
		persisted, _, err := repo.Find(ID)
		assert.NoError(t, err)

		replyAccount := &account.Account{
			Number:    reply.Number,
			OwnerName: reply.OwnerName,
		}
		assert.Empty(t, cmp.Diff(replyAccount, persisted))
	}
	shouldNotExist := func(t *testing.T, request *request, reply *reply) {
		ID, _ := uuid.FromBytes(request.Id)
		notPersisted, _, err := repo.Find(ID)
		assert.Nil(t, notPersisted)
		assert.IsType(t, &repositories.NotFoundError{}, err)
	}

	testCases := []testCase{
		{
			"valid",
			func() (context.Context, *request) {
				account := ValidAccount()
				bankID := uuid.New()
				ID, _ := repo.Add(account, bankID)
				ctx := AuthenticatedContext(context.Background(), bankID)
				return ctx, &request{Id: ID[:]}
			},
			codes.OK,
			shouldExist,
		},
		{
			"not found",
			func() (context.Context, *request) {
				missingID := uuid.New()
				bankID := uuid.New()
				ctx := AuthenticatedContext(context.Background(), bankID)
				return ctx, &request{Id: missingID[:]}
			},
			codes.NotFound,
			shouldNotExist,
		},
		{
			"unauthenticated",
			func() (context.Context, *request) {
				account := ValidAccount()
				bankID := uuid.New()
				ID, _ := repo.Add(account, bankID)
				return context.Background(), &request{Id: ID[:]}
			},
			codes.Unauthenticated,
			noPersistCheck,
		},
		{
			"permission denied",
			func() (context.Context, *request) {
				account := ValidAccount()
				ID, _ := repo.Add(account, uuid.New())
				bankID := uuid.New()
				ctx := AuthenticatedContext(context.Background(), bankID)
				return ctx, &request{Id: ID[:]}
			},
			codes.PermissionDenied,
			noPersistCheck,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			ctx, request := tc.request()

			reply, err := client.Find(ctx, request)

			if tc.status == codes.OK {
				assert.NotNil(t, reply)
			}
			status, _ := status.FromError(err)
			assert.Equal(t, tc.status, status.Code())

			tc.persistCheck(t, request, reply)
		})
	}
}
