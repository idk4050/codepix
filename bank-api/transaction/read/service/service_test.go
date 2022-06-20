package service_test

import (
	"codepix/bank-api/bankapitest"
	"codepix/bank-api/lib/repositories"
	pixkeyrepository "codepix/bank-api/pixkey/repository"
	proto "codepix/bank-api/proto/codepix/transaction/read"
	"codepix/bank-api/transaction"
	"codepix/bank-api/transaction/read/repository"
	"codepix/bank-api/transaction/transactiontest"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SenderIDs = transactiontest.SenderIDs

var ValidTransaction = transactiontest.ValidTransaction
var Service = transactiontest.ReadService
var ServiceWithMocks = transactiontest.ReadServiceWithMocks
var AuthenticatedContext = bankapitest.AuthenticatedContext

const projectionTimeout = time.Millisecond * 150
const projectionInterval = time.Millisecond * 50

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	client, repo, creator, tearDown := Service()
	defer tearDown()

	type test struct {
		description string
		fn          func(*testing.T)
	}
	tests := []test{
		{"find", FindIntegration(client, repo, creator)},
		{"list", ListIntegration(client, repo, creator)},
	}
	for i, test := range tests {
		t.Run(fmt.Sprint(i, "_", test.description), test.fn)
	}
}

func TestFind(t *testing.T) {
	type request = proto.FindRequest
	type reply = proto.FindReply
	type output = repository.Transaction

	type in struct {
		ctx     context.Context
		request *request
	}
	type out struct {
		output *output
		err    error
		reply  *reply
		status codes.Code
	}
	type testCase struct {
		description string
		in          in
		out         out
	}

	client, _, repo, _ := ServiceWithMocks()

	ID := uuid.New()

	valid := ValidTransaction()
	validRequest := &request{Id: ID[:]}

	validReply := &proto.FindReply{
		Id:           valid.ID[:],
		Sender:       valid.Sender[:],
		SenderBank:   valid.SenderBank[:],
		Receiver:     valid.Receiver[:],
		ReceiverBank: valid.ReceiverBank[:],

		CreatedAt:        timestamppb.New(valid.CreatedAt),
		UpdatedAt:        timestamppb.New(valid.UpdatedAt),
		Amount:           valid.Amount,
		Description:      valid.Description,
		Status:           proto.Status(valid.Status),
		ReasonForFailing: valid.ReasonForFailing,
	}

	senderCtx := AuthenticatedContext(context.Background(), valid.SenderBank)
	receiverCtx := AuthenticatedContext(context.Background(), valid.ReceiverBank)

	testCases := []testCase{
		{
			"valid as sender",
			in{
				senderCtx,
				validRequest,
			},
			out{
				valid,
				nil,
				validReply,
				codes.OK,
			},
		},
		{
			"valid as receiver",
			in{
				receiverCtx,
				validRequest,
			},
			out{
				valid,
				nil,
				validReply,
				codes.OK,
			},
		},
		{
			"not found",
			in{
				senderCtx,
				validRequest,
			},
			out{
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
				codes.Unauthenticated,
			},
		},
		{
			"permission denied",
			in{
				AuthenticatedContext(context.Background(), uuid.New()),
				validRequest,
			},
			out{
				valid,
				nil,
				nil,
				codes.PermissionDenied,
			},
		},
		{
			"internal error",
			in{
				senderCtx,
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
			if tc.out.status != codes.Unauthenticated {
				repo.On("Find", mock.IsType(tc.in.ctx), ID).Return(tc.out.output, tc.out.err).Once()
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

func FindIntegration(client proto.ServiceClient, repo repository.Repository, creator transactiontest.Creator,
) func(t *testing.T) {
	return func(t *testing.T) {
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

		noPersistCheck := func(t *testing.T, request *request, reply *reply) {}
		shouldExist := func(t *testing.T, request *request, reply *reply) {
			persisted, err := repo.Find(context.Background(), *(*uuid.UUID)(reply.Id))
			require.NoError(t, err)

			replyTransaction := &repository.Transaction{
				ID:           *(*uuid.UUID)(reply.Id),
				Sender:       *(*uuid.UUID)(reply.Sender),
				SenderBank:   *(*uuid.UUID)(reply.SenderBank),
				Receiver:     *(*uuid.UUID)(reply.Receiver),
				ReceiverBank: *(*uuid.UUID)(reply.ReceiverBank),

				CreatedAt:        reply.CreatedAt.AsTime(),
				UpdatedAt:        reply.UpdatedAt.AsTime(),
				Amount:           reply.Amount,
				Description:      reply.Description,
				Status:           transaction.Status(reply.Status),
				ReasonForFailing: reply.ReasonForFailing,
			}
			assert.Empty(t, cmp.Diff(replyTransaction, persisted))
		}
		shouldNotExist := func(t *testing.T, request *request, reply *reply) {
			_, err := repo.Find(context.Background(), *(*uuid.UUID)(request.Id))
			assert.IsType(t, &repositories.NotFoundError{}, err)
		}

		testCases := []testCase{
			{
				"valid as sender",
				func() (context.Context, *request) {
					ID, senderIDs, _ := creator.StartedIDs()
					ctx := AuthenticatedContext(context.Background(), senderIDs.BankID)
					return ctx, &request{Id: ID[:]}
				},
				codes.OK,
				shouldExist,
			},
			{
				"valid as receiver",
				func() (context.Context, *request) {
					ID, _, receiverIDs := creator.StartedIDs()
					ctx := AuthenticatedContext(context.Background(), receiverIDs.BankID)
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
					return context.Background(), &request{}
				},
				codes.Unauthenticated,
				noPersistCheck,
			},
			{
				"permission denied",
				func() (context.Context, *request) {
					ID, _, _ := creator.StartedIDs()
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

				if tc.status == codes.OK {
					assert.Eventually(t, func() bool {
						reply, _ := client.Find(ctx, request)
						return reply != nil
					}, projectionTimeout, projectionInterval)
				} else {
					assert.Never(t, func() bool {
						reply, _ := client.Find(ctx, request)
						return reply != nil
					}, projectionTimeout, projectionInterval)
				}

				reply, err := client.Find(ctx, request)
				status, _ := status.FromError(err)
				assert.Equal(t, tc.status.String(), status.Code().String())

				tc.persistCheck(t, request, reply)
			})
		}
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

	client, _, repo, _ := ServiceWithMocks()

	nTransactions := 4

	validOptions := repository.ListOptions{
		CreatedAfter: time.Now().UTC(),
		SenderID:     ValidTransaction().Sender,
		ReceiverID:   ValidTransaction().Receiver,
		Limit:        2,
		Skip:         1,
	}
	validRequest := &request{
		CreatedAfter: timestamppb.New(validOptions.CreatedAfter),
		SenderId:     validOptions.SenderID[:],
		ReceiverId:   validOptions.ReceiverID[:],
		Limit:        validOptions.Limit,
		Skip:         validOptions.Skip,
	}

	valid := []repository.ListItem{}
	for i := 0; i < nTransactions; i++ {
		valid = append(valid, *ValidTransaction())
	}

	validReplyItems := []*proto.ListItem{}
	for _, transaction := range valid {
		ID := transaction.ID
		Sender := transaction.Sender
		SenderBank := transaction.SenderBank
		Receiver := transaction.Receiver
		ReceiverBank := transaction.ReceiverBank

		validReplyItems = append(validReplyItems, &proto.ListItem{
			Id:           ID[:],
			Sender:       Sender[:],
			SenderBank:   SenderBank[:],
			Receiver:     Receiver[:],
			ReceiverBank: ReceiverBank[:],

			CreatedAt:        timestamppb.New(transaction.CreatedAt),
			UpdatedAt:        timestamppb.New(transaction.UpdatedAt),
			Amount:           transaction.Amount,
			Description:      transaction.Description,
			Status:           proto.Status(transaction.Status),
			ReasonForFailing: transaction.ReasonForFailing,
		})
	}
	validReply := &proto.ListReply{
		Items: validReplyItems,
	}

	senderCtx := AuthenticatedContext(context.Background(), valid[0].SenderBank)
	receiverCtx := AuthenticatedContext(context.Background(), valid[0].ReceiverBank)

	testCases := []testCase{
		{
			"valid as sender",
			in{
				senderCtx,
				validRequest,
			},
			out{
				valid,
				nil,
				validReply,
				codes.OK,
			},
		},
		{
			"valid as receiver",
			in{
				receiverCtx,
				validRequest,
			},
			out{
				valid,
				nil,
				validReply,
				codes.OK,
			},
		},
		{
			"not found",
			in{
				senderCtx,
				validRequest,
			},
			out{
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
				codes.Unauthenticated,
			},
		},
		{
			"permission denied",
			in{
				AuthenticatedContext(context.Background(), uuid.New()),
				validRequest,
			},
			out{
				valid,
				nil,
				nil,
				codes.PermissionDenied,
			},
		},
		{
			"internal error",
			in{
				senderCtx,
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
			if tc.out.status != codes.Unauthenticated {
				repo.On("List", mock.IsType(tc.in.ctx), validOptions).
					Return(tc.out.output, tc.out.err).Once()
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

func ListIntegration(client proto.ServiceClient, repo repository.Repository, creator transactiontest.Creator,
) func(t *testing.T) {
	return func(t *testing.T) {
		if testing.Short() {
			t.Skip()
		}
		type request = proto.ListRequest
		type reply = proto.ListReply

		type testCase struct {
			description   string
			nTransactions int
			request       func() (context.Context, *request)
			status        codes.Code
			persistCheck  func(t *testing.T, request *request, reply *reply, nTransactions int)
		}

		noPersistCheck := func(*testing.T, *request, *reply, int) {}

		shouldExist := func(t *testing.T, request *request, reply *reply, nTransactions int) {
			senderID, _ := uuid.FromBytes(request.SenderId)
			receiverID, _ := uuid.FromBytes(request.ReceiverId)

			transactions, err := repo.List(context.Background(), repository.ListOptions{
				CreatedAfter: request.CreatedAfter.AsTime(),
				SenderID:     senderID,
				ReceiverID:   receiverID,
				Limit:        request.Limit,
				Skip:         request.Skip,
			})
			require.NoError(t, err)
			require.Len(t, transactions, nTransactions)
			require.Len(t, reply.Items, nTransactions)

			for i := 0; i < nTransactions; i++ {
				replyItem := reply.Items[i]
				received := repository.ListItem{
					ID:               *(*uuid.UUID)(replyItem.Id),
					Sender:           *(*uuid.UUID)(replyItem.Sender),
					SenderBank:       *(*uuid.UUID)(replyItem.SenderBank),
					Receiver:         *(*uuid.UUID)(replyItem.Receiver),
					ReceiverBank:     *(*uuid.UUID)(replyItem.ReceiverBank),
					CreatedAt:        replyItem.CreatedAt.AsTime(),
					UpdatedAt:        replyItem.UpdatedAt.AsTime(),
					Amount:           replyItem.Amount,
					Description:      replyItem.Description,
					Status:           transaction.Status(replyItem.Status),
					ReasonForFailing: replyItem.ReasonForFailing,
				}
				assert.Empty(t, cmp.Diff(transactions[i], received))
			}
		}

		for i := 0; i < 5; i++ {
			go creator.StartedID(
				SenderIDs{AccountID: uuid.New(), BankID: uuid.New()},
				pixkeyrepository.IDs{AccountID: uuid.New(), BankID: uuid.New()},
			)
		}
		testCases := []testCase{
			{
				"valid as sender",
				3,
				func() (context.Context, *request) {
					senderID, senderBankID := uuid.New(), uuid.New()
					now := timestamppb.Now()

					for i := 0; i < 10; i++ {
						creator.StartedID(
							SenderIDs{AccountID: senderID, BankID: senderBankID},
							pixkeyrepository.IDs{AccountID: uuid.New(), BankID: uuid.New()},
						)
					}
					request := &request{
						CreatedAfter: now,
						SenderId:     senderID[:],
						Limit:        7,
						Skip:         7,
					}
					ctx := AuthenticatedContext(context.Background(), senderBankID)
					return ctx, request
				},
				codes.OK,
				shouldExist,
			},
			{
				"valid as receiver",
				3,
				func() (context.Context, *request) {
					receiverID, receiverBankID := uuid.New(), uuid.New()
					now := timestamppb.Now()

					for i := 0; i < 10; i++ {
						creator.StartedID(
							SenderIDs{AccountID: uuid.New(), BankID: uuid.New()},
							pixkeyrepository.IDs{AccountID: receiverID, BankID: receiverBankID},
						)
					}
					request := &request{
						CreatedAfter: now,
						ReceiverId:   receiverID[:],
						Limit:        7,
						Skip:         7,
					}
					ctx := AuthenticatedContext(context.Background(), receiverBankID)
					return ctx, request
				},
				codes.OK,
				shouldExist,
			},
			{
				"valid empty",
				0,
				func() (context.Context, *request) {
					senderID, senderBankID, receiverID := uuid.New(), uuid.New(), uuid.New()

					request := &request{
						CreatedAfter: &timestamppb.Timestamp{},
						SenderId:     senderID[:],
						ReceiverId:   receiverID[:],
						Limit:        10,
						Skip:         0,
					}
					ctx := AuthenticatedContext(context.Background(), senderBankID)
					return ctx, request
				},
				codes.OK,
				shouldExist,
			},
			{
				"unauthenticated",
				10,
				func() (context.Context, *request) {
					senderID, receiverID := uuid.New(), uuid.New()

					request := &request{
						CreatedAfter: &timestamppb.Timestamp{},
						SenderId:     senderID[:],
						ReceiverId:   receiverID[:],
						Limit:        10,
						Skip:         0,
					}
					return context.Background(), request
				},
				codes.Unauthenticated,
				noPersistCheck,
			},
			{
				"permission denied",
				4,
				func() (context.Context, *request) {
					senderID, senderBankID := uuid.New(), uuid.New()
					now := timestamppb.Now()

					for i := 0; i < 4; i++ {
						creator.StartedID(
							SenderIDs{AccountID: senderID, BankID: senderBankID},
							pixkeyrepository.IDs{AccountID: uuid.New(), BankID: uuid.New()},
						)
					}
					request := &request{
						CreatedAfter: now,
						SenderId:     senderID[:],
						Limit:        4,
						Skip:         0,
					}
					ctx := AuthenticatedContext(context.Background(), uuid.New())
					return ctx, request
				},
				codes.PermissionDenied,
				noPersistCheck,
			},
		}
		for i, tc := range testCases {
			t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
				ctx, request := tc.request()

				switch {
				case tc.nTransactions == 0:
					time.Sleep(projectionTimeout)

				case tc.status == codes.OK:
					assert.Eventually(t, func() bool {
						reply, _ := client.List(ctx, request)
						return reply != nil && len(reply.Items) >= tc.nTransactions
					}, projectionTimeout, projectionInterval)

				default:
					assert.Never(t, func() bool {
						reply, _ := client.List(ctx, request)
						return reply != nil
					}, projectionTimeout, projectionInterval)
				}

				reply, err := client.List(ctx, request)
				status, _ := status.FromError(err)
				assert.Equal(t, tc.status.String(), status.Code().String())

				tc.persistCheck(t, request, reply, tc.nTransactions)
			})
		}
	}
}
