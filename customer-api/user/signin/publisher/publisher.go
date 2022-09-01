package publisher

import (
	"codepix/customer-api/lib/outboxes"
	"codepix/customer-api/lib/publishers"
	"codepix/customer-api/user/signin/eventhandler"
)

type Publisher struct {
	EventHandler eventhandler.EventHandler
}

func (p Publisher) Publish(message outboxes.Message) error {
	switch message.Type {
	case eventhandler.Started{}.Type():
		event, err := publishers.Payload[eventhandler.Started](message)
		if err != nil {
			return err
		}
		return p.EventHandler.Started(event)
	default:
		return outboxes.NewUnknownMessageTypeError(message)
	}
}
