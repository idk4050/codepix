package service_test

import (
	rpc "codepix/bank-api/adapters/rpc"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/bankapitest"
	"codepix/bank-api/lib/repositories"
	"codepix/bank-api/pixkey"
	"codepix/bank-api/pixkey/pixkeytest"
	"codepix/bank-api/pixkey/repository"
	proto "codepix/bank-api/proto/codepix/pixkey"
	"context"
	"fmt"
	"testing"

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

var ValidPixKey = pixkeytest.ValidPixKey
var InvalidPixKey = pixkeytest.InvalidPixKey
var Service = pixkeytest.Service
var ServiceWithMocks = pixkeytest.ServiceWithMocks
var AuthenticatedContext = bankapitest.AuthenticatedContext

func TestRegister(t *testing.T) {
	type request = proto.RegisterRequest
	type reply = proto.RegisterReply

	type in struct {
		ctx     context.Context
		request *request
	}
	type out struct {
		ID     *uuid.UUID
		err    error
		reply  *reply
		status *status.Status
	}
	type testCase struct {
		description string
		in          in
		out         out
	}

	client, repo := ServiceWithMocks()

	valid := ValidPixKey()
	invalid := InvalidPixKey()

	ID := uuid.New()
	accountID, bankID := uuid.New(), uuid.New()

	ctx := AuthenticatedContext(context.Background(), bankID)
	ctxWithLocale := metadata.AppendToOutgoingContext(ctx, "locale", validator.EN_US)

	validRequest := &request{Type: proto.Type(valid.Type), Key: valid.Key, AccountId: accountID[:]}

	testCases := []testCase{
		{
			"valid",
			in{
				ctx,
				validRequest,
			},
			out{
				&ID,
				nil,
				&reply{Id: ID[:]},
				status.New(codes.OK, ""),
			},
		},
		{
			"invalid",
			in{
				ctxWithLocale,
				&request{Type: proto.Type(invalid.Type), Key: invalid.Key, AccountId: nil},
			},
			out{
				nil,
				nil,
				nil,
				func() *status.Status {
					status, _ := status.New(codes.InvalidArgument,
						"validation failed on type (required)").
						WithDetails(rpc.ValidationErrorMessage(map[string]string{
							"account_id": "account_id is a required field",
							"key":        "Key must be a maximum of 100 characters in length",
							"type":       "Key type is a required field",
						}))
					return status
				}(),
			},
		},
		{
			"already exists",
			in{
				ctx,
				validRequest,
			},
			out{
				nil,
				&repositories.AlreadyExistsError{},
				nil,
				status.New(codes.AlreadyExists, ""),
			},
		},
		{
			"unauthenticated",
			in{
				context.Background(),
				&request{},
			},
			out{
				nil,
				nil,
				nil,
				status.New(codes.Unauthenticated, ""),
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
				&repositories.InternalError{},
				nil,
				status.New(codes.Internal, ""),
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			if !(tc.out.ID == nil && tc.out.err == nil) {
				repo.On("Add", mock.IsType(pixkey.PixKey{}), accountID).
					Return(tc.out.ID, tc.out.err).Once()
			}

			reply, err := client.Register(tc.in.ctx, tc.in.request)

			status, _ := status.FromError(err)
			assert.Equal(t, tc.out.status.Code().String(), status.Code().String())

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
		request      func() (context.Context, *request)
		status       *status.Status
		persistCheck func(t *testing.T, request *request, reply *reply)
	}

	client, repo, _ := Service()

	noPersistCheck := func(t *testing.T, request *request, reply *reply) {}
	shouldPersist := func(t *testing.T, request *request, reply *reply) {
		ID, _ := uuid.FromBytes(reply.Id)
		pixKey, IDs, err := repo.Find(ID)
		assert.NoError(t, err)
		assert.EqualValues(t, request.Type, pixKey.Type)
		assert.Equal(t, request.Key, pixKey.Key)
		assert.Equal(t, ID, IDs.PixKeyID)
	}
	shouldNotPersist := func(t *testing.T, request *request, reply *reply) {
		_, _, err := repo.FindByKey(request.Key)
		assert.IsType(t, &repositories.NotFoundError{}, err)
	}

	validRequest := func() (context.Context, *request) {
		pk := ValidPixKey()
		accountID, bankID := uuid.New(), uuid.New()

		ctx := AuthenticatedContext(context.Background(), bankID)
		request := &request{
			Type:      proto.Type(pk.Type),
			Key:       pk.Key,
			AccountId: accountID[:],
		}
		return ctx, request
	}
	invalidRequest := func() (context.Context, *request) {
		pk := InvalidPixKey()
		bankID := uuid.New()

		ctx := AuthenticatedContext(context.Background(), bankID)
		ctx = metadata.AppendToOutgoingContext(ctx, "locale", validator.EN_US)
		request := &request{
			Type:      proto.Type(pk.Type),
			Key:       pk.Key,
			AccountId: nil,
		}
		return ctx, request
	}
	alreadyExistsRequest := func() (context.Context, *request) {
		pk := ValidPixKey()
		accountID, bankID := uuid.New(), uuid.New()
		repo.Add(pk, accountID, bankID)

		ctx := AuthenticatedContext(context.Background(), bankID)
		request := &request{
			Type:      proto.Type(pk.Type),
			Key:       pk.Key,
			AccountId: accountID[:],
		}
		return ctx, request
	}
	unauthenticatedRequest := func() (context.Context, *request) {
		return context.Background(), &request{}
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
					"validation failed on type (required)").
					WithDetails(rpc.ValidationErrorMessage(map[string]string{
						"account_id": "account_id is a required field",
						"key":        "Key must be a maximum of 100 characters in length",
						"type":       "Key type is a required field",
					}))
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
		{
			"unauthenticated",
			unauthenticatedRequest,
			status.New(codes.Unauthenticated, ""),
			shouldNotPersist,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			ctx, request := tc.request()

			reply, err := client.Register(ctx, request)

			status, _ := status.FromError(err)
			assert.Equal(t, tc.status.Code().String(), status.Code().String())

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
	type output = pixkey.PixKey

	type in struct {
		ctx     context.Context
		request *request
	}
	type out struct {
		output    *output
		outputIDs *repository.IDs
		err       error
		reply     *reply
		status    codes.Code
	}
	type testCase struct {
		description string
		in          in
		out         out
	}

	client, repo := ServiceWithMocks()

	ID := uuid.New()
	accountID := uuid.New()
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
				&repository.IDs{PixKeyID: ID, AccountID: accountID, BankID: bankID},
				nil,
				&reply{
					Id:        ID[:],
					Type:      proto.Type(valid.Type),
					Key:       valid.Key,
					AccountId: accountID[:],
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
			"permission denied",
			in{
				ctx,
				validRequest,
			},
			out{
				&valid,
				&repository.IDs{PixKeyID: ID, AccountID: accountID, BankID: uuid.New()},
				nil,
				nil,
				codes.PermissionDenied,
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
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			if !(tc.out.output == nil && tc.out.err == nil) {
				repo.On("Find", ID).Return(tc.out.output, tc.out.outputIDs, tc.out.err).Once()
			}

			reply, err := client.Find(tc.in.ctx, tc.in.request)

			status, _ := status.FromError(err)
			assert.Equal(t, tc.out.status.String(), status.Code().String())

			if tc.out.status == codes.OK {
				assert.Empty(t, cmp.Diff(tc.out.reply, reply, protocmp.Transform()))
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
		status       codes.Code
		persistCheck func(t *testing.T, request *request, reply *reply)
	}

	client, repo, creator := Service()

	noPersistCheck := func(t *testing.T, request *request, reply *reply) {}
	shouldExist := func(t *testing.T, request *request, reply *reply) {
		require.Equal(t, request.Id, reply.Id)

		persisted, IDs, err := repo.Find(*(*uuid.UUID)(reply.Id))
		require.NoError(t, err)
		assert.Equal(t, reply.Id, IDs.PixKeyID[:])

		expected := &pixkey.PixKey{
			Type: pixkey.Type(reply.Type),
			Key:  reply.Key,
		}
		assert.Empty(t, cmp.Diff(expected, persisted))
	}
	shouldNotExist := func(t *testing.T, request *request, reply *reply) {
		ID, _ := uuid.FromBytes(request.Id)
		pixKey, _, err := repo.Find(ID)
		assert.Nil(t, pixKey)
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
				pixKey := ValidPixKey()
				IDs := creator.PixKeyIDs(pixKey)
				return context.Background(), &request{Id: IDs.PixKeyID[:]}
			},
			codes.Unauthenticated,
			noPersistCheck,
		},
		{
			"permission denied",
			func() (context.Context, *request) {
				pixKey := ValidPixKey()
				IDs := creator.PixKeyIDs(pixKey)
				bankID := uuid.New()
				ctx := AuthenticatedContext(context.Background(), bankID)
				return ctx, &request{Id: IDs.PixKeyID[:]}
			},
			codes.PermissionDenied,
			noPersistCheck,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			ctx, request := tc.request()

			reply, err := client.Find(ctx, request)

			status, _ := status.FromError(err)
			assert.Equal(t, tc.status.String(), status.Code().String())

			if tc.status == codes.OK {
				assert.NotNil(t, reply)
			}
			tc.persistCheck(t, request, reply)
		})
	}
}

func TestList(t *testing.T) {
	type request = proto.ListRequest
	type reply = proto.ListReply
	type output = []repository.ListItem

	type in struct {
		ctx     context.Context
		request *request
	}
	type out struct {
		output output
		err    error
		reply  *reply
		status codes.Code
	}
	type testCase struct {
		description string
		in          in
		out         out
	}

	client, repo := ServiceWithMocks()

	accountID, bankID := uuid.New(), uuid.New()
	ctx := AuthenticatedContext(context.Background(), bankID)

	nPixKeys := 10
	valid := []repository.ListItem{}
	for i := 0; i < nPixKeys; i++ {
		pixKey := ValidPixKey()
		valid = append(valid, repository.ListItem{
			ID:   uuid.New(),
			Type: pixKey.Type,
			Key:  pixKey.Key,
		})
	}

	validReply := []*proto.ListItem{}
	for _, pixKey := range valid {
		ID := pixKey.ID
		validReply = append(validReply, &proto.ListItem{
			Id:   ID[:],
			Type: proto.Type(pixKey.Type),
			Key:  pixKey.Key,
		})
	}
	validRequest := &request{AccountId: accountID[:]}

	testCases := []testCase{
		{
			"valid",
			in{
				ctx,
				validRequest,
			},
			out{
				valid,
				nil,
				&reply{
					Items: validReply,
				},
				codes.OK,
			},
		},
		{
			"valid empty",
			in{
				ctx,
				validRequest,
			},
			out{
				[]repository.ListItem{},
				nil,
				&reply{
					Items: []*proto.ListItem{},
				},
				codes.OK,
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
				&repositories.InternalError{},
				nil,
				codes.Internal,
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			if !(tc.out.output == nil && tc.out.err == nil) {
				options := repository.ListOptions{
					AccountID: accountID,
					BankID:    bankID,
				}
				repo.On("List", options).Return(tc.out.output, tc.out.err).Once()
			}

			reply, err := client.List(tc.in.ctx, tc.in.request)

			status, _ := status.FromError(err)
			assert.Equal(t, tc.out.status.String(), status.Code().String())

			if tc.out.status == codes.OK {
				assert.Empty(t, cmp.Diff(tc.out.reply, reply, protocmp.Transform()))
			}
		})
	}
}

func TestListIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	type request = proto.ListRequest
	type reply = proto.ListReply

	type testCase struct {
		description  string
		request      func() (context.Context, *request, uuid.UUID)
		status       codes.Code
		persistCheck func(t *testing.T, request *request, reply *reply, bankID uuid.UUID)
	}

	client, repo, _ := Service()

	noPersistCheck := func(t *testing.T, request *request, reply *reply, bankID uuid.UUID) {}
	shouldExist := func(nPixKeys int) func(t *testing.T, request *request, reply *reply, bankID uuid.UUID) {
		return func(t *testing.T, request *request, reply *reply, bankID uuid.UUID) {
			accountID, _ := uuid.FromBytes(request.AccountId)
			pixKeys, err := repo.List(repository.ListOptions{
				AccountID: accountID,
				BankID:    bankID,
			})
			assert.NoError(t, err)
			require.Len(t, pixKeys, nPixKeys)
			require.Len(t, reply.Items, nPixKeys)

			for i := 0; i < nPixKeys; i++ {
				replyItem := reply.Items[i]
				received := repository.ListItem{
					ID:   *(*uuid.UUID)(replyItem.Id),
					Type: pixkey.Type(replyItem.Type),
					Key:  replyItem.Key,
				}
				assert.Empty(t, cmp.Diff(pixKeys[i], received))
			}
		}
	}

	for i := 0; i < 10; i++ {
		repo.Add(ValidPixKey(), uuid.New(), uuid.New())
	}
	testCases := []testCase{
		{
			"valid",
			func() (context.Context, *request, uuid.UUID) {
				accountID, bankID := uuid.New(), uuid.New()
				for i := 0; i < 10; i++ {
					repo.Add(ValidPixKey(), accountID, bankID)
				}
				ctx := AuthenticatedContext(context.Background(), bankID)
				return ctx, &request{AccountId: accountID[:]}, bankID
			},
			codes.OK,
			shouldExist(10),
		},
		{
			"valid empty",
			func() (context.Context, *request, uuid.UUID) {
				accountID, bankID := uuid.New(), uuid.New()
				ctx := AuthenticatedContext(context.Background(), bankID)
				return ctx, &request{AccountId: accountID[:]}, bankID
			},
			codes.OK,
			shouldExist(0),
		},
		{
			"unauthenticated",
			func() (context.Context, *request, uuid.UUID) {
				accountID := uuid.New()
				return context.Background(), &request{AccountId: accountID[:]}, uuid.New()
			},
			codes.Unauthenticated,
			noPersistCheck,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			ctx, request, bankID := tc.request()

			reply, err := client.List(ctx, request)

			status, _ := status.FromError(err)
			assert.Equal(t, tc.status.String(), status.Code().String())

			if tc.status == codes.OK {
				assert.NotNil(t, reply)
			}
			tc.persistCheck(t, request, reply, bankID)
		})
	}
}
