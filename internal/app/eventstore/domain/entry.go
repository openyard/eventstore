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
	e := &Event{
		id:          id,
		name:        name,
		aggregateID: aggregateID,
		occurredAt:  occurredAt,
		payload:     payload,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
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
		"Name":          e.name,
		"ID":            e.id,
		"AggregateID":   e.aggregateID,
		"CorrelationID": e.correlationID,
		"CausationID":   e.causationID,
		"ContentType":   e.contentType,
		"OccurredAt":    e.occurredAt,
		"Meta":          make(map[string]any, len(e.meta)),
		"Payload":       e.payload,
	}
	for k, val := range e.meta {
		v["Meta"].(map[string]any)[k] = val
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
	if v["CorrelationID"] != nil {
		e.correlationID = v["CorrelationID"].(string)
	}
	if v["CausationID"] != nil {
		e.causationID = v["CausationID"].(string)
	}
	if v["ContentType"] != nil {
		e.contentType = v["ContentType"].(string)
	}
	e.occurredAt, _ = time.Parse(time.RFC3339Nano, v["OccurredAt"].(string))
	if v["Meta"] != nil {
		meta := v["Meta"].(map[string]any)
		e.meta = make(map[string]string, len(meta))
		for k, v := range meta {
			e.meta[k] = v.(string)
		}
	}
	if v["Payload"] != nil {
		e.payload = []byte(v["Payload"].(string))
	}
	return nil
}
