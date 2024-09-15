package domain

import (
	"encoding/json"
	"time"
)

type Entry struct {
	GlobalPos uint64
	Stream    string
	StreamPos uint64
	Event     *Event
}

type Event struct {
	id          string
	name        string
	aggregateID string
	occurredAt  time.Time
	payload     []byte
}

func NewEventAt(id, name, aggregateID string, occurredAt time.Time, payload []byte) *Event {
	return &Event{
		id:          id,
		name:        name,
		aggregateID: aggregateID,
		occurredAt:  occurredAt,
		payload:     payload,
	}
}

func (e *Event) ID() string {
	return e.id
}

func (e *Event) Name() string {
	return e.name
}

func (e *Event) AggregateID() string {
	return e.aggregateID
}

func (e *Event) Payload() []byte {
	return e.payload
}

func (e *Event) OccurredAt() time.Time {
	return e.occurredAt
}

// MarshalJSON is implementation of json.Marshaler
func (e *Event) MarshalJSON() ([]byte, error) {
	v := map[string]any{
		"Name":        e.name,
		"ID":          e.id,
		"AggregateID": e.aggregateID,
		"Payload":     e.payload,
		"OccurredAt":  e.occurredAt,
	}
	return json.MarshalIndent(v, "", "  ")
}

// UnmarshalJSON is implementation of json.Unmarshaler
func (e *Event) UnmarshalJSON(data []byte) error {
	var v map[string]any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	e.name = v["Name"].(string)
	e.id = v["ID"].(string)
	e.aggregateID = v["AggregateID"].(string)
	if v["Payload"] != nil {
		e.payload = []byte(v["Payload"].(string))
	}
	e.occurredAt, _ = time.Parse(time.RFC3339Nano, v["OccurredAt"].(string))
	return nil
}
