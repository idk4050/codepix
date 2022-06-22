package outbox

import (
	"codepix/customer-api/adapters/databaseclient"
	"codepix/customer-api/config"
	"codepix/customer-api/lib/outboxes"
	"codepix/customer-api/lib/publishers"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/omaskery/outboxen-gorm/pkg/storage"
	"github.com/omaskery/outboxen/pkg/outbox"
	"gorm.io/gorm"
)

type Outbox struct {
	config        outbox.Config
	inner         *outbox.Outbox
	storage       *storage.Storage
	database      *gorm.DB
	writeLogger   logr.Logger
	processLogger logr.Logger
	storageLogger logr.Logger
}

func New(config config.Config, logger logr.Logger, publishers map[outboxes.Namespace]publishers.Publisher,
) (*Outbox, error) {
	log := logger.WithName("outbox")
	storageLogger := log.WithName("storage")

	database, err := databaseclient.Open(config, storageLogger)
	if err != nil {
		return nil, err
	}
	storage := storage.New(database)
	storage.IDGenerator = &idGenerator{}

	publisherAdapter := publisherAdapter{
		Publishers: publishers,
		Logger:     log.WithName("publisher"),
	}
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	outboxConfig := outbox.Config{
		Storage:         storage,
		Publisher:       publisherAdapter,
		ProcessorID:     hostname,
		Logger:          log,
		Clock:           clockwork.NewRealClock(),
		ProcessInterval: outbox.DefaultProcessInterval,
	}
	outbox, err := outbox.New(outboxConfig)
	if err != nil {
		return nil, err
	}
	gormOutbox := &Outbox{
		config:        outboxConfig,
		inner:         outbox,
		storage:       storage,
		database:      database,
		writeLogger:   log.WithName("writer"),
		processLogger: log.WithName("processor"),
		storageLogger: storageLogger,
	}
	return gormOutbox, nil
}

func (o Outbox) AutoMigrate() error {
	err := o.storage.AutoMigrate()
	o.database.Logger = databaseclient.NewLogger(o.storageLogger.V(1))
	return err
}

func (o Outbox) Start(ctx context.Context) {
	o.processLogger.Info("processor started")
	defer o.processLogger.Info("processor stopped")

	for {
		select {
		case <-ctx.Done():
			o.processLogger.Info("context cancelled")
			return
		case <-o.config.Clock.After(o.config.ProcessInterval):
			o.processLogger.V(1).Info("woken by processing interval")
		}

		op := func() error {
			return o.inner.PumpOutbox(ctx)
		}
		notify := func(err error, duration time.Duration) {
			o.processLogger.V(1).Error(err, "pump error (transient)", "backoff", duration)
		}
		bo := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
		if err := backoff.RetryNotify(op, bo, notify); err != nil {
			o.processLogger.Error(err, "pump error (giving up)")
		}
	}
}

type idGenerator struct{}

func (u *idGenerator) GenerateID(_ clockwork.Clock, message outbox.Message) string {
	_, id := getTypeAndID(message)
	return id
}

func getTypeAndID(message outbox.Message) (string, string) {
	parts := strings.Split(string(message.Key), idSeparator)
	return parts[0], parts[1]
}

const idSeparator = "///"

func (o Outbox) Write(tx interface{}, message outboxes.NewMessage) error {
	namespace := message.Namespace()
	ctx := outbox.WithNamespace(context.Background(), namespace)

	messageType := message.Type()
	if strings.Contains(messageType, idSeparator) {
		return fmt.Errorf("message type cannot contain '%s'", idSeparator)
	}

	id := uuid.NewString()
	key := fmt.Sprintf("%s%s%s", messageType, idSeparator, id)

	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	msg := &outbox.Message{
		Key:     []byte(key),
		Payload: payload,
	}

	kvs := []any{
		"namespace", namespace,
		"type", messageType,
		"id", id,
	}
	err = o.inner.Publish(ctx, tx, *msg)

	if err == nil {
		o.writeLogger.Info("message written", kvs...)
	} else {
		o.writeLogger.Error(err, "message not written", kvs...)
	}
	return err
}
