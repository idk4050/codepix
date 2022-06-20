package service_test

import (
	"codepix/bank-api/account/accounttest"
	rpcproto "codepix/bank-api/adapters/rpc/proto"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/bankapitest"
	"codepix/bank-api/lib/repositories"
	"codepix/bank-api/pixkey"
	"codepix/bank-api/pixkey/interactor"
	"codepix/bank-api/pixkey/pixkeytest"
	"codepix/bank-api/pixkey/service/proto"
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
var ValidPixKey = pixkeytest.ValidPixKey
var InvalidPixKey = pixkeytest.InvalidPixKey
var Service = pixkeytest.Service
var ServiceWithMocks = pixkeytest.ServiceWithMocks
var AuthenticatedContext = bankapitest.AuthenticatedContext

func TestRegister(t *testing.T) {
	type request = proto.RegisterRequest
	type reply = proto.RegisterReply
	type input = interactor.RegisterInput
	type output = interactor.RegisterOutput

	type in struct {
		ctx     context.Context
		request *request
		input   *input
	}
	type out struct {
		permissionErr error
		output        *output
		err           error
		reply         *reply
		status        *status.Status
	}
	type testCase struct {
		description string
		in          in
		out         out
	}

	client, interactor, _, accountRepo := ServiceWithMocks()

	valid := ValidPixKey()
	invalid := InvalidPixKey()

	ID := uuid.New()
	bankID := uuid.New()
	accountID := uuid.New()

	ctx := AuthenticatedContext(context.Background(), bankID)
	ctxWithLocale := metadata.AppendToOutgoingContext(ctx, "locale", validator.EN_US)

	validInput := &input{Type: valid.Type, Key: valid.Key, AccountID: accountID}
	validRequest := &request{Type: proto.Type(valid.Type), Key: valid.Key, AccountId: accountID[:]}

	testCases := []testCase{
		{
			"valid",
			in{
				ctx,
				validRequest,
				validInput,
			},
			out{
				nil,
				&output{PixKey: valid, ID: ID},
				nil,
				&reply{Id: ID[:]},
				nil,
			},
		},
		{
			"invalid",
			in{
				ctxWithLocale,
				&request{Type: proto.Type(invalid.Type), Key: invalid.Key, AccountId: nil},
				nil,
			},
			out{
				nil,
				nil,
				nil,
				nil,
				func() *status.Status {
					status, _ := status.New(codes.InvalidArgument,
						"validation failed on type (required)").
						WithDetails(&rpcproto.ValidationError{Errors: map[string]string{
							"account_id": "Account is a required field",
							"key":        "Key is a required field",
							"type":       "Key type is a required field",
						}})
					return status
				}(),
			},
		},
		{
			"account not found",
			in{
				ctx,
				validRequest,
				validInput,
			},
			out{
				nil,
				nil,
				repositories.NewNotFoundError("account"),
				nil,
				status.New(codes.NotFound, "account not found"),
			},
		},
		{
			"permission denied (not the account's owner)",
			in{
				ctx,
				validRequest,
				nil,
			},
			out{
				repositories.NewNotFoundError("account"),
				nil,
				nil,
				nil,
				status.New(codes.PermissionDenied, "account not found"),
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
				nil,
				repositories.NewAlreadyExistsError("pix key"),
				nil,
				status.New(codes.AlreadyExists, "pix key already exists"),
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
				nil,
				repositories.NewInternalError("insert", "pix key", "error message"),
				nil,
				status.New(codes.Internal, "internal error"),
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
				nil,
				errors.New("error message"),
				nil,
				status.New(codes.Unknown, "unknown error"),
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			if tc.in.request != nil {
				accountID, _ := uuid.FromBytes(tc.in.request.AccountId)
				accountRepo.On("ExistsWithBankID", accountID).Return(tc.out.permissionErr).Once()
			}
			if tc.in.input != nil {
				interactor.On("Register", *tc.in.input).Return(tc.out.output, tc.out.err).Once()
			}

			reply, err := client.Register(tc.in.ctx, tc.in.request)

			if tc.out.status == nil {
				assert.NotNil(t, reply)
				assert.NoError(t, err)
				assert.Empty(t, cmp.Diff(tc.out.reply, reply, protocmp.Transform()))
			} else {
				assert.Nil(t, reply)
				status, _ := status.FromError(err)
				assert.JSONEq(t, protojson.Format(tc.out.status.Proto()), protojson.Format(status.Proto()))
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
		request      func() (context.Context, *request)
		status       *status.Status
		persistCheck func(t *testing.T, request *request, reply *reply)
	}

	client, repo, creator := Service()

	validRequest := func() (context.Context, *request) {
		pk := ValidPixKey()
		accountIDs := creator.AccountIDs(ValidAccount())

		ctx := AuthenticatedContext(context.Background(), accountIDs.BankID)
		request := &request{Type: proto.Type(pk.Type), Key: pk.Key, AccountId: accountIDs.AccountID[:]}
		return ctx, request
	}
	permissionDeniedRequest := func() (context.Context, *request) {
		pk := ValidPixKey()
		accountIDs := creator.AccountIDs(ValidAccount())

		ctx := AuthenticatedContext(context.Background(), uuid.New())
		request := &request{Type: proto.Type(pk.Type), Key: pk.Key, AccountId: accountIDs.AccountID[:]}
		return ctx, request
	}
	invalidRequest := func() (context.Context, *request) {
		pk := InvalidPixKey()
		accountIDs := creator.AccountIDs(ValidAccount())

		ctx := AuthenticatedContext(context.Background(), accountIDs.BankID)
		ctxWithLocale := metadata.AppendToOutgoingContext(ctx, "locale", validator.EN_US)

		request := &request{Type: proto.Type(pk.Type), Key: pk.Key, AccountId: nil}
		return ctxWithLocale, request
	}
	alreadyExistsRequest := func() (context.Context, *request) {
		pk := ValidPixKey()
		accountIDs := creator.AccountIDs(ValidAccount())
		repo.Add(pk, accountIDs.AccountID)

		ctx := AuthenticatedContext(context.Background(), accountIDs.BankID)
		request := &request{Type: proto.Type(pk.Type), Key: pk.Key, AccountId: accountIDs.AccountID[:]}
		return ctx, request
	}

	noPersistCheck := func(t *testing.T, request *request, reply *reply) {}
	shouldPersist := func(t *testing.T, request *request, reply *reply) {
		requestPixKey := pixkey.PixKey{Type: pixkey.Type(request.Type), Key: request.Key}

		ID, _ := uuid.FromBytes(reply.Id)
		pixKey, persistedIDs, err := repo.Find(ID)
		assert.NoError(t, err)
		assert.Empty(t, cmp.Diff(*pixKey, requestPixKey))
		assert.Equal(t, ID, persistedIDs.PixKeyID)
	}
	shouldNotPersist := func(t *testing.T, request *request, reply *reply) {
		notPersisted, _, err := repo.FindByKey(request.Key)
		assert.Nil(t, notPersisted)
		assert.IsType(t, &repositories.NotFoundError{}, err)
	}

	testCases := []testCase{
		{
			"valid",
			validRequest,
			nil,
			shouldPersist,
		},
		{
			"permission denied (not the account's owner)",
			permissionDeniedRequest,
			status.New(codes.PermissionDenied, "account not found"),
			shouldNotPersist,
		},
		{
			"invalid",
			invalidRequest,
			func() *status.Status {
				status, _ := status.New(codes.InvalidArgument,
					"validation failed on type (required)").
					WithDetails(&rpcproto.ValidationError{Errors: map[string]string{
						"account_id": "Account is a required field",
						"key":        "Key is a required field",
						"type":       "Key type is a required field",
					}})
				return status
			}(),
			shouldNotPersist,
		},
		{
			"already exists",
			alreadyExistsRequest,
			status.New(codes.AlreadyExists, "pix key already exists"),
			noPersistCheck,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			ctx, request := tc.request()

			reply, err := client.Register(ctx, request)

			if tc.status == nil {
				assert.NotNil(t, reply)
				assert.NoError(t, err)
			} else {
				assert.Nil(t, reply)
				status, _ := status.FromError(err)
				assert.JSONEq(t, protojson.Format(tc.status.Proto()), protojson.Format(status.Proto()))
			}
			tc.persistCheck(t, request, reply)
		})
	}
}

func TestFind(t *testing.T) {
	type request = proto.FindRequest
	type reply = proto.FindReply
	type output = pixkey.PixKey

	type in struct {
		ctx     context.Context
		request *request
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

	client, _, repo, _ := ServiceWithMocks()

	ID := uuid.New()
	bankID := uuid.New()

	ctx := AuthenticatedContext(context.Background(), bankID)

	valid := ValidPixKey()
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
				nil,
				&proto.FindReply{
					Type: proto.Type(valid.Type),
					Key:  valid.Key,
				},
				nil,
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
				repositories.NewNotFoundError("pix key"),
				nil,
				status.New(codes.NotFound, "pix key not found"),
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
				repositories.NewInternalError("select", "pix key", "error message"),
				nil,
				status.New(codes.Internal, "internal error"),
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
				errors.New("error message"),
				nil,
				status.New(codes.Unknown, "unknown error"),
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			repo.On("Find", ID).Return(tc.out.output, nil, tc.out.err).Once()

			reply, err := client.Find(tc.in.ctx, tc.in.request)

			if tc.out.status == nil {
				assert.NotNil(t, reply)
				assert.NoError(t, err)
				assert.Empty(t, cmp.Diff(tc.out.reply, reply, protocmp.Transform()))
			} else {
				assert.Nil(t, reply)
				status, _ := status.FromError(err)
				assert.JSONEq(t, protojson.Format(tc.out.status.Proto()), protojson.Format(status.Proto()))
			}
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
		status       *status.Status
		persistCheck func(t *testing.T, request *request, reply *reply)
	}

	client, repo, creator := Service()

	shouldExist := func(t *testing.T, request *request, reply *reply) {
		ID, _ := uuid.FromBytes(request.Id)
		persisted, _, err := repo.Find(ID)
		assert.NoError(t, err)

		replyPixKey := &pixkey.PixKey{
			Type: pixkey.Type(reply.Type),
			Key:  reply.Key,
		}
		assert.Empty(t, cmp.Diff(replyPixKey, persisted))
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
				pixKey := ValidPixKey()
				IDs := creator.PixKeyIDs(pixKey)

				ctx := AuthenticatedContext(context.Background(), IDs.BankID)
				return ctx, &request{Id: IDs.PixKeyID[:]}
			},
			nil,
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
			status.New(codes.NotFound, "pix key not found"),
			shouldNotExist,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			ctx, request := tc.request()

			reply, err := client.Find(ctx, request)

			if tc.status == nil {
				assert.NotNil(t, reply)
				assert.NoError(t, err)
			} else {
				assert.Nil(t, reply)
				status, _ := status.FromError(err)
				assert.JSONEq(t, protojson.Format(tc.status.Proto()), protojson.Format(status.Proto()))
			}
			tc.persistCheck(t, request, reply)
		})
	}
}
