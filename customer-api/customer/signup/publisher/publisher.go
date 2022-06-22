package publisher

import (
	"codepix/customer-api/customer/signup/eventhandler"
	"codepix/customer-api/lib/outboxes"
	"codepix/customer-api/lib/publishers"
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
	case eventhandler.Finished{}.Type():
		event, err := publishers.Payload[eventhandler.Finished](message)
		if err != nil {
			return err
		}
		return p.EventHandler.Finished(event)
	default:
		return outboxes.NewUnknownMessageTypeError(message)
	}
}
