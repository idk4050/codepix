package outbox

import (
	"codepix/customer-api/lib/outboxes"
	"codepix/customer-api/lib/publishers"
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/omaskery/outboxen/pkg/outbox"
)

type publisherAdapter struct {
	Publishers map[outboxes.Namespace]publishers.Publisher
	Logger     logr.Logger
}

func (p publisherAdapter) Publish(ctx context.Context, msgs ...outbox.Message) error {
	namespace := outbox.NamespaceFromContext(ctx)
	publisher, ok := p.Publishers[namespace]
	if !ok {
		return fmt.Errorf("publisher of %s namespace not found", namespace)
	}

	errors := []error{}
	atLeastOneError := false
	for _, msg := range msgs {
		message := convertMessage(msg)

		kvs := []any{
			"namespace", namespace,
			"type", message.Type,
			"id", message.ID,
		}
		err := publisher.Publish(message)

		if err == nil {
			p.Logger.Info("message published", kvs...)
			errors = append(errors, nil)
		} else if _, skip := err.(*publishers.SkipMessage); skip {
			errors = append(errors, nil)
		} else {
			p.Logger.Error(err, "message not published", kvs...)
			errors = append(errors, publishers.NewPublishError(err, message, namespace))
			atLeastOneError = true
		}
	}
	if atLeastOneError {
		return &outbox.PublishError{Errors: errors}
	}
	return nil
}

func convertMessage(message outbox.Message) outboxes.Message {
	typ, id := getTypeAndID(message)
	return outboxes.Message{
		ID:      id,
		Type:    typ,
		Payload: message.Payload,
	}
}
