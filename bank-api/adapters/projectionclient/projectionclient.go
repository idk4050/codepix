package projectionclient

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"codepix/bank-api/adapters/eventhandler"
	"codepix/bank-api/config"

	"github.com/go-logr/logr"
	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/eventhandler/projector"
	"github.com/looplab/eventhorizon/repo/mongodb"
	"github.com/phayes/freeport"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type StoreProjection struct {
	projectionName string
	client         *mongo.Client
	outbox         eventhorizon.Outbox
	logger         logr.Logger
	onClose        func() error
}

func Open(ctx context.Context, config config.Config, logger logr.Logger, outbox eventhorizon.Outbox,
) (*StoreProjection, error) {
	cfg := config.StoreProjection
	logger = logger.WithName("projection")

	var onClose func() error
	var URI string

	opts := options.Client().
		SetWriteConcern(writeconcern.New(writeconcern.WMajority())).
		SetReadConcern(readconcern.Majority()).
		SetReadPreference(readpref.PrimaryPreferred())

	if cfg.InMemory {
		freePort, err := freeport.GetFreePort()
		if err != nil {
			return nil, fmt.Errorf("open store projection: %w", err)
		}
		port := fmt.Sprint(freePort)

		container, err := exec.Command(
			"podman", "run", "--detach", "--rm", "-p", port+":"+port,
			"docker.io/mongo:6.0.1", "/bin/sh", "-c",
			fmt.Sprintf(`
			port="%s"
			( for i in $(seq 30); do
					mongosh --port $port --quiet --eval "rs.initiate({ _id: 'rs0', members: [
						{ _id: 0, host: 'localhost:$port' }
					]})"
					ok=$(mongosh --port $port --quiet --eval "!!rs.isMaster().primary")
					[ "$ok" = "true" ] && break
					sleep 0.5
				done
				[ "$ok" != "true" ] && mongod --shutdown
			) &
			mongod --bind_ip_all --port $port --replSet rs0
			`, port),
		).Output()
		if err != nil {
			return nil, fmt.Errorf("open store projection: %w", err)
		}
		containerID := strings.TrimSpace(string(container))
		onClose = func() error {
			return exec.Command("podman", "container", "kill", containerID).Run()
		}
		URI = fmt.Sprintf("mongodb://localhost:%s/%s?replicaSet=rs0", port, cfg.Name)
	} else {
		onClose = func() error {
			return nil
		}
		URI = fmt.Sprintf("mongodb://%s:%s@%s/%s?replicaSet=%s",
			cfg.User, cfg.Password, strings.Join(cfg.Hosts, ","), cfg.Name, cfg.ReplicaSetName)
	}
	client, err := mongo.Connect(ctx, opts.ApplyURI(URI))
	if err != nil {
		return nil, fmt.Errorf("open event store: %w", err)
	}
	logger.Info("store projection opened")

	storeProjection := &StoreProjection{
		projectionName: cfg.Name,
		client:         client,
		outbox:         outbox,
		logger:         logger,
		onClose:        onClose,
	}
	return storeProjection, nil
}

func (sp *StoreProjection) Close() error {
	err := sp.client.Disconnect(context.Background())
	if err != nil {
		sp.logger.Error(err, "store projection failed to close")
		return err
	}
	err = sp.onClose()
	if err != nil {
		sp.logger.Error(err, "store projection failed to close")
		return err
	}
	sp.logger.Info("store projection closed")
	return nil
}

func (sp *StoreProjection) Setup(
	projectionType projector.Type,
	entity func() eventhorizon.Entity,
	entityProjector projector.Projector,
	aggregate eventhorizon.AggregateType,
) (*mongodb.Repo, error) {
	repo, err := mongodb.NewRepoWithClient(
		sp.client,
		sp.projectionName,
		string(projectionType),
		mongodb.WithConnectionCheck(nil),
	)
	if err != nil {
		return nil, fmt.Errorf("start %s projection: %w", projectionType, err)
	}
	repo.SetEntityFactory(entity)

	projectorHandler := projector.NewEventHandler(
		entityProjector,
		repo,
		projector.WithRetryOnce(),
	)
	projectorHandler.SetEntityFactory(entity)

	err = sp.outbox.AddHandler(context.Background(),
		eventhorizon.MatchAggregates{aggregate},
		wrappedHandler{
			eventhandler.Logger(sp.logger, projectorHandler),
			aggregate,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("start %s projection: %w", projectionType, err)
	}
	sp.logger.Info(fmt.Sprintf("%s projection started", projectionType))
	return repo, nil
}

type wrappedHandler struct {
	handler   eventhorizon.EventHandler
	aggregate eventhorizon.AggregateType
}

func (w wrappedHandler) HandlerType() eventhorizon.EventHandlerType {
	return eventhorizon.EventHandlerType(w.aggregate + "_projection")
}

func (w wrappedHandler) HandleEvent(ctx context.Context, event eventhorizon.Event) error {
	return w.handler.HandleEvent(ctx, event)
}
