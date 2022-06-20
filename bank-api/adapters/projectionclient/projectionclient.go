package projectionclient

import (
	"context"
	"fmt"
	"strings"

	"codepix/bank-api/config"

	"github.com/go-logr/logr"
	"github.com/looplab/eventhorizon"
	projectorhandler "github.com/looplab/eventhorizon/eventhandler/projector"
	"github.com/looplab/eventhorizon/repo/cache"
	mongorepo "github.com/looplab/eventhorizon/repo/mongodb"
	"github.com/tryvium-travels/memongo"
	"github.com/tryvium-travels/memongo/memongolog"
)

type StoreProjection struct {
	Connect      Connect
	OnDisconnect OnDisconnect
}

type Connect func(
	projectionType string,
	entityType func() eventhorizon.Entity,
	projector projectorhandler.Projector,
	matcher eventhorizon.EventMatcher,
) (eventhorizon.ReadWriteRepo, error)
type OnDisconnect func() error

func Open(config config.Config, logger logr.Logger, outbox eventhorizon.Outbox,
) (*StoreProjection, error) {
	cfg := config.StoreProjection
	log := logger.WithName("projection")

	var projectionURI string
	var projectionName string
	var onDisconnect OnDisconnect

	if cfg.InMemory {
		server, err := memongo.StartWithOptions(
			&memongo.Options{
				ShouldUseReplica: true,
				LogLevel:         memongolog.LogLevelWarn,
				MongodBin:        cfg.InMemoryBinaryPath,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("open store projection client: %w", err)
		}

		projectionURI = server.URI()
		projectionName = fmt.Sprintf("%s_%s", cfg.Name, memongo.RandomDatabase())

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

		projectionName = cfg.Name
		projectionURI = fmt.Sprintf("mongodb://%s%s/%s?replicaSet=%s",
			credentials, hostsAndPorts, projectionName, cfg.ReplicaSetName)

		onDisconnect = func() error {
			return nil
		}
	}

	connect := func(
		projectionType string,
		entityType func() eventhorizon.Entity,
		projector projectorhandler.Projector,
		matcher eventhorizon.EventMatcher,
	) (eventhorizon.ReadWriteRepo, error,
	) {
		repo, err := mongorepo.NewRepo(
			projectionURI,
			projectionName,
			projectionType,
			mongorepo.WithConnectionCheck(nil),
		)
		if err != nil {
			return nil, fmt.Errorf("open store projection client: %w", err)
		}
		repo.SetEntityFactory(entityType)

		cachedProjection := cache.NewRepo(repo)

		projectorHandler := projectorhandler.NewEventHandler(
			projector,
			cachedProjection,
			projectorhandler.WithRetryOnce(),
		)
		projectorHandler.SetEntityFactory(entityType)

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
			if err := projectorHandler.HandleEvent(ctx, event); err != nil {
				log.Error(err, "event not projected", kvs...)
				return err
			}
			log.Info("event projected", kvs...)
			return nil
		}

		err = outbox.AddHandler(context.Background(), matcher, handler)
		if err != nil {
			return nil, fmt.Errorf("open store projection client: %w", err)
		}
		return cachedProjection, nil
	}
	return &StoreProjection{
		connect,
		onDisconnect,
	}, nil
}
