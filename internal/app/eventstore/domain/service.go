package domain

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"sort"

	"github.com/openyard/eventstore/internal/app/kvstore"
)

type ServiceOpts func(*Service)

type Service struct {
	kvs kvstore.KeyValueStore
}

func NewService(opts ...ServiceOpts) *Service {
	s := &Service{kvs: &noopKVs{}}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithKeyValueStore(kvs kvstore.KeyValueStore) ServiceOpts {
	return func(s *Service) {
		s.kvs = kvs
	}
}

func (s *Service) HandleFunc(cmd Command) error {
	switch cmd.kind {
	case AppendCmd:
		return s.append(cmd.ctx, cmd.payload.(AppendCommand))
	case SubscribeCmd:
		// ...
		return nil
	case SubscribeWithOffsetCmd:
		// ...
		return nil
	default:
		return fmt.Errorf("unknown command: <%v>", cmd.kind)
	}
}

func (s *Service) QueryFunc(cmd Command) ([]Stream, error) {
	switch cmd.kind {
	case ReadCmd:
		return s.read(cmd.ctx, cmd.payload.(ReadCommand))
	case ReadAtCmd:
		return s.readAt(cmd.ctx, cmd.payload.(ReadAtCommand))
	default:
		return nil, fmt.Errorf("unknown command: <%v>", cmd.kind)
	}
}

func (s *Service) append(_ context.Context, cmd AppendCommand) error {
	if err := s.kvs.WithTx(func() error {
		for _, streamData := range cmd.streamData {
			version, err := s.kvs.Get(kvstore.KvsBucketIndex, streamData.name)
			if err != nil && streamData.expectedVersion > 0 {
				log.Printf("[ERROR]\t %T.append - read index failed: %s", s, err)
				return err
			}
			if binary.BigEndian.Uint64(version) != streamData.expectedVersion {
				log.Printf("[ERROR]\t %T.append - concurrent write mismatch index: %d(actual) != %d(expected)", s, binary.BigEndian.Uint64(version), streamData.expectedVersion)
				return fmt.Errorf("[%4d] concurrent write mismatch index: %d(actual) != %d(expected)",
					ErrConcurrentChange, binary.BigEndian.Uint64(version), streamData.expectedVersion)
			}
			var stream *Stream
			if streamData.expectedVersion == 0 {
				stream = s.firstAppend(streamData)
			} else if stream, err = s.nextAppend(streamData); err != nil {
				return err
			}

			log.Printf("[DEBUG]\t %T.append - entries: %+v", s, stream.events)
			var raw []byte
			raw, err = stream.MarshalJSON()
			if err != nil {
				log.Printf("[ERROR]\t %T.append - marshaling error: %s", s, err)
				return err
			}
			log.Printf("[TRACE]\t %T.append - raw stream: %s", s, string(raw))
			idx := make([]byte, 8)
			binary.BigEndian.PutUint64(idx, stream.version)
			if err := s.kvs.Put(kvstore.KvsBucketContent, stream.name, raw); err != nil {
				return err
			}
			if err := s.kvs.Put(kvstore.KvsBucketIndex, stream.name, idx); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		s.kvs.Rollback()
		return err
	}
	return nil
}

func (s *Service) read(_ context.Context, cmd ReadCommand) ([]Stream, error) {
	result := make([]Stream, 0)
	for _, stream := range cmd.streams {
		streamData, err := s.kvs.Get(kvstore.KvsBucketContent, stream)
		if err != nil {
			return result, err
		}
		var elem Stream
		if err := elem.UnmarshalJSON(streamData); err != nil {
			return result, err
		}
		if uint64(len(elem.events)) != elem.version {
			return result, fmt.Errorf("[ERROR]\t %T.read !!! version mismatch in stream <%s>: version=%d, events=%d",
				s, elem.name, elem.version, len(elem.events))
		}
		result = append(result, elem)
	}
	return result, nil
}

func (s *Service) readAt(_ context.Context, cmd ReadAtCommand) ([]Stream, error) {
	result := make([]Stream, 0)
	for _, stream := range cmd.streams {
		streamData, err := s.kvs.Get(kvstore.KvsBucketContent, stream)
		if err != nil {
			return result, err
		}
		var elem Stream
		if err := elem.UnmarshalJSON(streamData); err != nil {
			return result, err
		}
		if uint64(len(elem.events)) != elem.version {
			return result, fmt.Errorf("[ERROR]\t %T.readAt !!! version mismatch in stream <%s>: version=%d, events=%d",
				s, elem.name, elem.version, len(elem.events))
		}
		ev := make(map[uint64]*Event)
		for i, e := range elem.events {
			if !e.OccurredAt().After(cmd.at) {
				ev[i] = e
			}
		}
		elem.events = ev
		result = append(result, elem)
	}
	return result, nil
}

func (s *Service) firstAppend(streamData StreamData) *Stream {
	sort.Slice(streamData.events, func(i, j int) bool {
		return streamData.events[i].OccurredAt().Before(streamData.events[j].OccurredAt())
	})
	entries := make(map[uint64]*Event, len(streamData.events))
	for idx, e := range streamData.events {
		entries[uint64(idx)] = e
	}
	return buildStream(streamData.name, uint64(len(entries)), entries)
}

func (s *Service) nextAppend(streamData StreamData) (*Stream, error) {
	raw, err := s.kvs.Get(kvstore.KvsBucketContent, streamData.name)
	if err != nil && streamData.expectedVersion > 0 {
		return nil, err
	}
	var stream Stream
	if err = stream.UnmarshalJSON(raw); err != nil {
		return nil, err
	}
	log.Printf("current stream: %s", &stream)
	for idx, e := range streamData.events {
		stream.events[stream.version+uint64(idx)] = e
	}
	stream.version += uint64(len(streamData.events))
	return &stream, nil
}
