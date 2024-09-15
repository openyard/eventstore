package domain

import (
	"reflect"
	"testing"
	"time"
)

type args struct {
	name   string
	events map[uint64]*Event
}

type fields struct {
	name    string
	version uint64
	events  map[uint64]*Event
}

var (
	ts1, _     = time.Parse(time.RFC3339Nano, "2024-09-11T21:51:39.810959063Z")
	ts2, _     = time.Parse(time.RFC3339Nano, "2024-09-11T21:51:39.810967123Z")
	ts3, _     = time.Parse(time.RFC3339Nano, "2024-09-11T21:51:39.810968821Z")
	testEvents = map[uint64]*Event{
		1: NewEventAt("cb1801f0-54f9-4928-8f5d-edf8000270d3", "event-1", "id", ts1, nil),
		2: NewEventAt("77b8c58e-779b-4a3a-846c-e70947889c07", "event-2", "id", ts2, nil),
		3: NewEventAt("08a7ac82-6843-4970-983e-1e8c32edad2a", "event-3", "id", ts3, nil),
	}
)

func TestBuildStream(t *testing.T) {
	tests := []struct {
		name string
		args args
		want *Stream
	}{
		{
			"build stream",
			args{
				"test-stream",
				testEvents,
			},
			&Stream{
				name:    "test-stream",
				version: 3,
				events:  testEvents,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildStream(tt.args.name, 0, tt.args.events); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewStream(t *testing.T) {
	tests := []struct {
		name string
		args args
		want *Stream
	}{
		{
			"build stream",
			args{
				"test-stream",
				nil,
			},
			&Stream{
				name:    "test-stream",
				version: 0,
				events:  make(map[uint64]*Event),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStream(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			"marshal stream",
			fields{
				"test-stream",
				3,
				testEvents,
			},
			[]byte(`{
  "Events": {
    "1": {
      "AggregateID": "id",
      "ID": "cb1801f0-54f9-4928-8f5d-edf8000270d3",
      "Name": "event-1",
      "OccurredAt": "2024-09-11T21:51:39.810959063Z",
      "Payload": null
    },
    "2": {
      "AggregateID": "id",
      "ID": "77b8c58e-779b-4a3a-846c-e70947889c07",
      "Name": "event-2",
      "OccurredAt": "2024-09-11T21:51:39.810967123Z",
      "Payload": null
    },
    "3": {
      "AggregateID": "id",
      "ID": "08a7ac82-6843-4970-983e-1e8c32edad2a",
      "Name": "event-3",
      "OccurredAt": "2024-09-11T21:51:39.810968821Z",
      "Payload": null
    }
  },
  "Name": "test-stream",
  "Version": 3
}`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stream{
				name:    tt.fields.name,
				version: tt.fields.version,
				events:  tt.fields.events,
			}
			got, err := s.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestStream_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"marshal stream",
			fields{
				"test-stream",
				0,
				nil,
			},
			args{
				[]byte(`{
  "Events": {
    "1": {
      "AggregateID": "id",
      "ID": "cb1801f0-54f9-4928-8f5d-edf8000270d3",
      "Kind": "",
      "Name": "event-1",
      "OccurredAt": "2024-09-11T21:51:39.810959063Z",
      "Payload": null
    },
    "2": {
      "AggregateID": "id",
      "ID": "77b8c58e-779b-4a3a-846c-e70947889c07",
      "Kind": "",
      "Name": "event-2",
      "OccurredAt": "2024-09-11T21:51:39.810967123Z",
      "Payload": null
    },
    "3": {
      "AggregateID": "id",
      "ID": "08a7ac82-6843-4970-983e-1e8c32edad2a",
      "Kind": "",
      "Name": "event-3",
      "OccurredAt": "2024-09-11T21:51:39.810968821Z",
      "Payload": null
    }
  },
  "Name": "test-stream",
  "Version": 3
}`),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stream{
				name:    tt.fields.name,
				version: tt.fields.version,
				events:  tt.fields.events,
			}
			if err := s.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if "08a7ac82-6843-4970-983e-1e8c32edad2a" != s.events[3].ID() {
				t.Errorf("Unexpected EventID got = %s, want %s", s.events[3].ID(), "08a7ac82-6843-4970-983e-1e8c32edad2a")
			}
		})
	}
}
