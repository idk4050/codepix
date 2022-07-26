package eventbus

import (
	"codepix/bank-api/adapters/eventhandler"
	"codepix/bank-api/config"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-redis/redis/v9"
	"github.com/looplab/eventhorizon"
	"github.com/phayes/freeport"
)

type EventBus struct {
	config  config.Config
	client  *redis.Client
	outbox  eventhorizon.Outbox
	onClose func() error
	logger  logr.Logger
}

func Open(ctx context.Context, config config.Config, logger logr.Logger, outbox eventhorizon.Outbox,
) (*EventBus, error) {
	cfg := config.EventBus
	logger = logger.WithName("eventbus")

	var clientOpts *redis.Options
	var onClose func() error

	if cfg.InMemory {
		freePort, err := freeport.GetFreePort()
		if err != nil {
			return nil, fmt.Errorf("open event bus: %w", err)
		}
		port := fmt.Sprint(freePort)

		container, err := exec.Command(
			"podman", "run", "--detach", "--rm", "-p", port+":6379", "docker.io/redis:7.0.4",
		).Output()
		if err != nil {
			return nil, fmt.Errorf("open event bus: %w", err)
		}
		containerID := strings.TrimSpace(string(container))

		clientOpts = &redis.Options{
			Addr: fmt.Sprintf("%s:%s", "localhost", port),
		}
		onClose = func() error {
			return exec.Command("podman", "stop", containerID).Run()
		}
	} else {
		clientOpts = &redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			Username: cfg.User,
			Password: cfg.Password,
		}
		onClose = func() error {
			return nil
		}
	}
	client := redis.NewClient(clientOpts)
	if res, err := client.Ping(ctx).Result(); err != nil || res != "PONG" {
		return nil, fmt.Errorf("open event bus: %w", err)
	}
	logger.Info("event bus opened")

	eventBus := &EventBus{
		config:  config,
		client:  client,
		outbox:  outbox,
		onClose: onClose,
		logger:  logger,
	}
	return eventBus, nil
}

func (b *EventBus) Close() error {
	err := b.client.Close()
	if err != nil {
		b.logger.Error(err, "event bus failed to close")
		return err
	}
	err = b.onClose()
	if err != nil {
		b.logger.Error(err, "event bus failed to close")
		return err
	}
	b.logger.Info("event bus closed")
	return nil
}

const eventKey = "event"

func (b *EventBus) CreateReader(blockDuration, maxPendingAge time.Duration) (*Reader, error) {
	return &Reader{
		Client:        b.client,
		BlockDuration: blockDuration,
		MaxPendingAge: maxPendingAge,
	}, nil
}

func (b *EventBus) SetupWriter(eventType eventhorizon.EventType,
	streams func(eventhorizon.Event) []string,
) error {
	writer := &Writer{b.client, streams}

	err := b.outbox.AddHandler(context.Background(),
		eventhorizon.MatchEvents{eventType},
		wrappedWriter{
			eventhandler.Logger(b.logger, writer),
			eventType,
		},
	)
	if err != nil {
		return fmt.Errorf("setup %s writer: %w", eventType, err)
	}
	b.logger.Info(fmt.Sprintf("%s writer setup", eventType))
	return nil
}

type wrappedWriter struct {
	handler   eventhorizon.EventHandler
	eventType eventhorizon.EventType
}

func (w wrappedWriter) HandlerType() eventhorizon.EventHandlerType {
	return eventhorizon.EventHandlerType(w.eventType + "_eventbus_writer")
}

func (w wrappedWriter) HandleEvent(ctx context.Context, event eventhorizon.Event) error {
	return w.handler.HandleEvent(ctx, event)
}
