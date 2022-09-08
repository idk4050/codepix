package stream

import (
	"codepix/example-bank-api/adapters/pix"
	readproto "codepix/example-bank-api/proto/codepix/transaction/read"
	writeproto "codepix/example-bank-api/proto/codepix/transaction/write"
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	protobuf "google.golang.org/protobuf/proto"
)

func Run(ctx context.Context, logger logr.Logger, pixClient *pix.Client,
) {
	logger = logger.WithName("pix.tx.stream")
	readClient := readproto.NewStreamClient(pixClient.Conn)
	writeClient := writeproto.NewStreamClient(pixClient.Conn)

	go RunStarted(ctx, logger.WithName("started"), readClient, writeClient)
	go RunConfirmed(ctx, logger.WithName("confirmed"), readClient, writeClient)
	go RunCompleted(ctx, logger.WithName("completed"), readClient)
	go RunFailed(ctx, logger.WithName("failed"), readClient)
}

const retrySleep = time.Second * 5

func RunStream[R, W grpc.ClientStream, T any](ctx context.Context, logger logr.Logger,
	makeReadClient func(ctx context.Context, opts ...grpc.CallOption) (R, error),
	getIDs func(*T) []uuid.UUID,
	makeWriteClient func(ctx context.Context, opts ...grpc.CallOption) (W, error),
	respond func(ID uuid.UUID, index int, msg *T) protobuf.Message,
) {
	for {
		readClient, err := makeReadClient(ctx)
		if err != nil {
			logger.Error(err, "failed to connect")
			time.Sleep(retrySleep)
			continue
		}
		var writeClient W
		if makeWriteClient != nil {
			writeClient, err = makeWriteClient(ctx)
			if err != nil {
				logger.Error(err, "failed to connect")
				time.Sleep(retrySleep)
				continue
			}
		}
		logger.Info("connected")
	connectionLoop:
		for {
			received := new(T)
			err := readClient.RecvMsg(received)
			if err != nil {
				logger.Error(err, "failed to receive")
				if status.Code(err) == codes.Unavailable {
					break connectionLoop
				}
				time.Sleep(retrySleep)
				continue
			}
			IDs := getIDs(received)
			logger.Info("received events", "ids", IDs)

			nacks := make([]bool, len(IDs))
			if makeWriteClient != nil {
				for i, ID := range IDs {
					err := writeClient.SendMsg(respond(ID, i, received))
					if err != nil {
						logger.Error(err, "failed to respond", "id", ID)
						if status.Code(err) == codes.Unavailable {
							break connectionLoop
						}
						nacks[i] = true
					}
				}
			}
			for {
				err := readClient.SendMsg(&readproto.Ack{
					Nacks: nacks,
				})
				if err != nil {
					logger.Error(err, "failed to ack", "ids", IDs, "nacks", nacks)
					if status.Code(err) == codes.Unavailable {
						break connectionLoop
					}
					time.Sleep(retrySleep)
					continue
				}
				logger.Info("acked events", "ids", IDs, "nacks", nacks)
				break
			}
		}
	}
}

func RunStarted(ctx context.Context, logger logr.Logger,
	readClient readproto.StreamClient, writeClient writeproto.StreamClient,
) {
	RunStream(ctx, logger,
		readClient.Started,
		func(msg *readproto.StartedTransactions) []uuid.UUID {
			IDs := []uuid.UUID{}
			for _, tx := range msg.Events {
				ID, _ := uuid.FromBytes(tx.Id)
				IDs = append(IDs, ID)
			}
			return IDs
		},
		writeClient.Confirm,
		func(ID uuid.UUID, index int, msg *readproto.StartedTransactions) protobuf.Message {
			return &writeproto.ConfirmRequest{
				Id: ID[:],
			}
		},
	)
}

func RunConfirmed(ctx context.Context, logger logr.Logger,
	readClient readproto.StreamClient, writeClient writeproto.StreamClient,
) {
	RunStream(ctx, logger,
		readClient.Confirmed,
		func(msg *readproto.ConfirmedTransactions) []uuid.UUID {
			IDs := []uuid.UUID{}
			for _, tx := range msg.Events {
				ID, _ := uuid.FromBytes(tx.Id)
				IDs = append(IDs, ID)
			}
			return IDs
		},
		writeClient.Complete,
		func(ID uuid.UUID, index int, msg *readproto.ConfirmedTransactions) protobuf.Message {
			return &writeproto.CompleteRequest{
				Id: ID[:],
			}
		},
	)
}

func RunCompleted(ctx context.Context, logger logr.Logger, readClient readproto.StreamClient) {
	RunStream[readproto.Stream_CompletedClient, grpc.ClientStream](
		ctx, logger,
		readClient.Completed,
		func(msg *readproto.ConfirmedTransactions) []uuid.UUID {
			IDs := []uuid.UUID{}
			for _, tx := range msg.Events {
				ID, _ := uuid.FromBytes(tx.Id)
				IDs = append(IDs, ID)
			}
			return IDs
		},
		nil,
		nil,
	)
}

func RunFailed(ctx context.Context, logger logr.Logger, readClient readproto.StreamClient) {
	RunStream[readproto.Stream_FailedClient, grpc.ClientStream](
		ctx, logger,
		readClient.Failed,
		func(msg *readproto.FailedTransactions) []uuid.UUID {
			IDs := []uuid.UUID{}
			for _, tx := range msg.Events {
				ID, _ := uuid.FromBytes(tx.Id)
				IDs = append(IDs, ID)
			}
			return IDs
		},
		nil,
		nil,
	)
}
