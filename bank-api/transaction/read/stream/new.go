package stream

import (
	"codepix/bank-api/adapters/eventbus"
	"codepix/bank-api/config"
	proto "codepix/bank-api/proto/codepix/transaction/read"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
)

func Register(server *grpc.Server, config config.Config, logger logr.Logger,
	eventBus *eventbus.EventBus) error {
	cfg := config.Transaction

	busReader, err := eventBus.CreateReader(cfg.BusBlockDuration, cfg.BusMaxPendingAge)
	if err != nil {
		return err
	}
	stream := &Stream{
		Logger:    logger.WithName("eventstream"),
		BusReader: busReader,
	}
	proto.RegisterStreamServer(server, stream)
	return nil
}
