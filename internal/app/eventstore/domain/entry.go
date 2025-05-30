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

/*
message Event {
  string ID = 1;
  string Name = 2;
  string AggregateID = 3;
  string CorrelationID = 4;
  string CausationID = 5;
  string ContentType = 6;
  google.protobuf.Timestamp OccurredAt = 7;
  map<string, string> Meta = 8;
  bytes Payload = 9;
}
*/

type Event struct {
	id            string
	name          string
	aggregateID   string
	correlationID string
	causationID   string
	contentType   string
	occurredAt    time.Time
	meta          map[string]string
	payload       []byte
}

type EventOpt func(e *Event)

func NewEventAt(id, name, aggregateID string, occurredAt time.Time, payload []byte, opts ...EventOpt) *Event {
	return &Event{
		id:          id,
		name:        name,
		aggregateID: aggregateID,
		occurredAt:  occurredAt,
		payload:     payload,
	}
}

func WithCorrelationID(correlationID string) EventOpt {
	return func(e *Event) {
		e.correlationID = correlationID
	}
}

func WithCausationID(causationID string) EventOpt {
	return func(e *Event) {
		e.causationID = causationID
	}
}

func WithContentType(contentType string) EventOpt {
	return func(e *Event) {
		e.contentType = contentType
	}
}

func WithMeta(meta map[string]string) EventOpt {
	return func(e *Event) {
		e.meta = meta
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

func (e *Event) CorrelationID() string {
	return e.correlationID
}

func (e *Event) CausationID() string {
	return e.causationID
}

func (e *Event) ContentType() string {
	return e.contentType
}

func (e *Event) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *Event) Meta() map[string]string {
	return e.meta
}

func (e *Event) Payload() []byte {
	return e.payload
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
