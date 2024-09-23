package domain

import (
	"encoding/json"
	"strconv"
)

type Stream struct {
	name    string
	version uint64
	events  map[uint64]*Event
}

// NewStream returns a new and empty stream in version 0
func NewStream(name string) *Stream {
	return &Stream{name, 0, make(map[uint64]*Event)}
}

func (s *Stream) Name() string {
	return s.name
}

func (s *Stream) Events() map[uint64]*Event {
	return s.events
}

// buildStream builds a stream from provided events and applies the current version and name to it
func buildStream(name string, version uint64, events map[uint64]*Event) *Stream {
	return &Stream{
		name:    name,
		version: version,
		events:  events,
	}
}

func (s *Stream) MarshalJSON() ([]byte, error) {
	v := map[string]any{
		"Name":    s.name,
		"Version": s.version,
		"Events":  s.events,
	}
	return json.MarshalIndent(v, "", "  ")
}

func (s *Stream) UnmarshalJSON(data []byte) error {
	var v map[string]any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	s.name = v["Name"].(string)
	s.version = uint64(v["Version"].(float64))
	var smap map[string]any
	smap = v["Events"].(map[string]any)
	s.events = make(map[uint64]*Event, len(smap))
	for k, v := range smap {
		if ki, err := strconv.ParseUint(k, 10, 64); err != nil {
			return err
		} else {
			var e Event
			if b, err := json.MarshalIndent(v, "", ""); err != nil {
				return err
			} else if err := e.UnmarshalJSON(b); err != nil {
				return err
			}
			s.events[ki] = &e
		}
	}
	return nil
}

func (s *Stream) String() string {
	b, _ := s.MarshalJSON()
	return string(b)
}
