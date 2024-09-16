package domain

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"sort"
)

type ServiceOpts func(*Service)

type Service struct {
	kvs KeyValueStore
}

func NewService(opts ...ServiceOpts) *Service {
	s := &Service{kvs: &noopKV{}}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithKeyValueStore(kvs KeyValueStore) ServiceOpts {
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
			version, err := s.kvs.Get(KvsBucketIndex, streamData.name)
			if err != nil && streamData.expectedVersion > 0 {
				log.Printf("[ERROR]\t %T.append - read index failed: %s", s, err)
				return err
			}
			if binary.BigEndian.Uint64(version) != streamData.expectedVersion {
				log.Printf("[ERROR]\t %T.append - concurrent write mismatch index: %d(actual) != %d(expected)", s, binary.BigEndian.Uint64(version), streamData.expectedVersion)
				return err
			}
			entries := make(map[uint64]*Event, len(streamData.events))
			sort.Slice(streamData.events, func(i, j int) bool {
				return streamData.events[i].OccurredAt().Before(streamData.events[j].OccurredAt())
			})
			for idx, e := range streamData.events {
				entries[uint64(idx)] = e
			}
			log.Printf("[DEBUG]\t %T.append - entries: %+v", s, entries)
			stream := buildStream(streamData.name, binary.BigEndian.Uint64(version), entries)
			raw, err := stream.MarshalJSON()
			if err != nil {
				log.Printf("[ERROR]\t %T.append - marshaling error: %s", s, err)
				return err
			}
			log.Printf("[TRACE]\t %T.append - raw stream: %s", s, string(raw))
			idx := make([]byte, 8)
			binary.BigEndian.PutUint64(idx, stream.version)
			if err := s.kvs.Put(KvsBucketContent, stream.name, raw); err != nil {
				return err
			}
			if err := s.kvs.Put(KvsBucketIndex, stream.name, idx); err != nil {
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
		streamData, err := s.kvs.Get(KvsBucketContent, stream)
		if err != nil {
			return result, err
		}
		var elem Stream
		if err := elem.UnmarshalJSON(streamData); err != nil {
			return result, err
		}
		result = append(result, elem)
	}
	return result, nil
}

func (s *Service) readAt(_ context.Context, cmd ReadAtCommand) ([]Stream, error) {
	result := make([]Stream, 0)
	for _, stream := range cmd.streams {
		streamData, err := s.kvs.Get(KvsBucketContent, stream)
		if err != nil {
			return result, err
		}
		var elem Stream
		if err := elem.UnmarshalJSON(streamData); err != nil {
			return result, err
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
