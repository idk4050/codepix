package eventstore

import (
	"codepix/bank-api/config"
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/looplab/eventhorizon"
	mongostore "github.com/looplab/eventhorizon/eventstore/mongodb"
	mongooutbox "github.com/looplab/eventhorizon/outbox/mongodb"
	"github.com/tryvium-travels/memongo"
	"github.com/tryvium-travels/memongo/memongolog"
)

type EventStore struct {
	Store        eventhorizon.EventStore
	Outbox       eventhorizon.Outbox
	OnDisconnect OnDisconnect
}

type OnDisconnect func() error

func Open(config config.Config, logger logr.Logger) (*EventStore, error) {
	cfg := config.EventStore
	log := logger.WithName("eventstore")

	var storeURI string
	var storeName string
	var onDisconnect OnDisconnect

	if cfg.InMemory {
		server, err := memongo.StartWithOptions(
			&memongo.Options{
				ShouldUseReplica: true,
				LogLevel:         memongolog.LogLevelInfo,
				MongodBin:        cfg.InMemoryBinaryPath,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("open event store: %w", err)
		}

		storeURI = server.URI()
		storeName = fmt.Sprintf("%s_%s", cfg.Name, memongo.RandomDatabase())

		onDisconnect = func() error {
			server.Stop()
			return nil
		}
	} else {
		password := cfg.Password
		if cfg.PasswordFromFile != "" {
			password = cfg.PasswordFromFile
		}
		credentials := fmt.Sprintf("%s:%s@", cfg.User, password)
		hosts := strings.Split(cfg.Host, ",")
		ports := strings.Split(cfg.Port, ",")

		var hostsAndPorts string
		for i, host := range hosts {
			hostsAndPorts += fmt.Sprintf("%s:%s", host, ports[i])
			if i < len(hosts)-1 {
				hostsAndPorts += ","
			}
		}

		storeName = cfg.Name
		storeURI = fmt.Sprintf("mongodb://%s%s/%s?replicaSet=%s",
			credentials, hostsAndPorts, storeName, cfg.ReplicaSetName)

		onDisconnect = func() error {
			return nil
		}
	}

	outbox, err := mongooutbox.NewOutbox(storeURI, storeName)
	if err != nil {
		return nil, fmt.Errorf("open event store outbox: %w", err)
	}

	var handler eventhorizon.EventHandlerFunc = func(
		ctx context.Context, event eventhorizon.Event,
	) error {
		kvs := []any{
			"aggregate", event.AggregateType(),
			"type", event.EventType(),
			"version", event.Version(),
			"created-at", event.Timestamp(),
			"id", event.AggregateID(),
		}
		if err := outbox.HandleEvent(ctx, event); err != nil {
			log.Error(err, "event not saved", kvs...)
			return err
		}
		log.Info("event saved", kvs...)
		return nil
	}

	store, err := mongostore.NewEventStoreWithClient(
		outbox.Client(),
		storeName,
		mongostore.WithEventHandlerInTX(handler),
	)
	if err != nil {
		return nil, fmt.Errorf("open event store: %w", err)
	}
	return &EventStore{
		Store:        store,
		Outbox:       outbox,
		OnDisconnect: onDisconnect,
	}, nil
}
