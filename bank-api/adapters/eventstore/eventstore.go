package eventstore

import (
	"codepix/bank-api/adapters/eventhandler"
	"codepix/bank-api/config"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
	"github.com/looplab/eventhorizon"
	mongostore "github.com/looplab/eventhorizon/eventstore/mongodb"
	mongooutbox "github.com/looplab/eventhorizon/outbox/mongodb"
	"github.com/phayes/freeport"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type EventStore struct {
	Store   eventhorizon.EventStore
	Outbox  eventhorizon.Outbox
	logger  logr.Logger
	onClose func() error
}

func Open(ctx context.Context, config config.Config, logger logr.Logger) (*EventStore, error) {
	cfg := config.EventStore
	logger = logger.WithName("eventstore")

	var onClose func() error
	var URI string

	opts := options.Client().
		SetWriteConcern(writeconcern.New(writeconcern.WMajority())).
		SetReadConcern(readconcern.Majority()).
		SetReadPreference(readpref.PrimaryPreferred())

	if cfg.InMemory {
		freePort, err := freeport.GetFreePort()
		if err != nil {
			return nil, fmt.Errorf("open event store: %w", err)
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
			return nil, fmt.Errorf("open event store: %w", err)
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
	outbox, err := mongooutbox.NewOutboxWithClient(client, cfg.Name)
	if err != nil {
		return nil, fmt.Errorf("open event store outbox: %w", err)
	}
	handler := wrappedHandler{
		eventhandler.Logger(logger, outbox),
	}
	store, err := mongostore.NewEventStoreWithClient(
		outbox.Client(),
		cfg.Name,
		mongostore.WithEventHandlerInTX(handler),
	)
	if err != nil {
		return nil, fmt.Errorf("open event store: %w", err)
	}
	logger.Info("event store opened")

	eventStore := &EventStore{
		Store:   store,
		Outbox:  outbox,
		logger:  logger,
		onClose: onClose,
	}
	return eventStore, nil
}

func (s *EventStore) Close() error {
	err := s.Outbox.Close()
	if err != nil {
		s.logger.Error(err, "event store failed to close")
		return err
	}
	err = s.onClose()
	if err != nil {
		s.logger.Error(err, "event store failed to close")
		return err
	}
	s.logger.Info("event store closed")
	return nil
}

func (s *EventStore) Start() error {
	s.Outbox.Start()
	s.logger.Info("event store started")
	return nil
}

type wrappedHandler struct {
	handler eventhorizon.EventHandler
}

func (wrappedHandler) HandlerType() eventhorizon.EventHandlerType { return "eventstore" }

func (w wrappedHandler) HandleEvent(ctx context.Context, event eventhorizon.Event) error {
	return w.handler.HandleEvent(ctx, event)
}
