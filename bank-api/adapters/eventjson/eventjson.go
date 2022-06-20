package eventjson

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
)

type Event struct {
	Type          eventhorizon.EventType     `json:"type"`
	Data          json.RawMessage            `json:"data"`
	Timestamp     time.Time                  `json:"timestamp"`
	AggregateType eventhorizon.AggregateType `json:"aggregate_type"`
	AggregateID   uuid.UUID                  `json:"aggregate_id"`
	Version       int                        `json:"version"`
	Metadata      map[string]interface{}     `json:"metadata"`
}

func Marshal(event eventhorizon.Event) ([]byte, error) {
	evt := Event{
		Type:          event.EventType(),
		Timestamp:     event.Timestamp(),
		AggregateType: event.AggregateType(),
		AggregateID:   event.AggregateID(),
		Version:       event.Version(),
		Metadata:      event.Metadata(),
	}
	if event.Data() != nil {
		data, err := json.Marshal(event.Data())
		if err != nil {
			return nil, fmt.Errorf("marshal event data: %w", err)
		}
		evt.Data = data
	}
	bytes, err := json.Marshal(evt)
	if err != nil {
		return nil, fmt.Errorf("marshal event: %w", err)
	}
	return bytes, nil
}

func Unmarshal(bytes []byte) (eventhorizon.Event, error) {
	var evt Event
	if err := json.Unmarshal(bytes, &evt); err != nil {
		return nil, fmt.Errorf("unmarshal event: %w", err)
	}
	data, err := eventhorizon.CreateEventData(evt.Type)
	if err != nil {
		return nil, fmt.Errorf("create event data: %w", err)
	}
	if err := json.Unmarshal(evt.Data, data); err != nil {
		return nil, fmt.Errorf("unmarshal event data: %w", err)
	}
	event := eventhorizon.NewEvent(
		evt.Type,
		data,
		evt.Timestamp,
		eventhorizon.ForAggregate(
			evt.AggregateType,
			evt.AggregateID,
			evt.Version,
		),
		eventhorizon.WithMetadata(evt.Metadata),
	)
	return event, nil
}
