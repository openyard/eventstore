package domain

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/openyard/eventstore/internal/app/persistance"
)

var _ Service = (*KVSService)(nil)

type KVSServiceOpts func(*KVSService)

type KVSService struct {
	kvs persistance.KeyValueStore
}

func NewKVSService(opts ...KVSServiceOpts) *KVSService {
	s := &KVSService{
		kvs: &noopKVS{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithKeyValueStore(kvs persistance.KeyValueStore) KVSServiceOpts {
	return func(s *KVSService) {
		s.kvs = kvs
	}
}

func (s *KVSService) HandleFunc(cmd Command) error {
	switch cmd.kind {
	case AppendCmd:
		return s.append(cmd.ctx, cmd.payload.(AppendCommand))
	default:
		return fmt.Errorf("unknown command: <%v>", cmd.kind)
	}
}

func (s *KVSService) QueryFunc(cmd Command) ([]Stream, error) {
	switch cmd.kind {
	case ReadCmd:
		return s.read(cmd.ctx, cmd.payload.(ReadCommand))
	case ReadAtCmd:
		return s.readAt(cmd.ctx, cmd.payload.(ReadAtCommand))
	default:
		return nil, fmt.Errorf("unknown command: <%v>", cmd.kind)
	}
}

func (s *KVSService) SubscribeFunc(cmd Command) ([]Entry, error) {
	switch cmd.kind {
	case SubscribeCmd:
		// ...
		return nil, fmt.Errorf("not implemented yet: <%v>", cmd.kind)
	case SubscribeWithOffsetCmd:
		// ...
		return nil, fmt.Errorf("not implemented yet: <%v>", cmd.kind)
	default:
		return nil, fmt.Errorf("unknown command: <%v>", cmd.kind)
	}
}

func (s *KVSService) append(_ context.Context, cmd AppendCommand) error {
	if err := s.kvs.WithTx(func() error {
		pos, err := s.readLogIndex()
		if err != nil {
			return err
		}
		for _, streamData := range cmd.streamData {
			version, err := s.kvs.Get(persistance.KvsBucketIndex, streamData.name)
			if err != nil && (streamData.expectedVersion > 0 || version == nil) {
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

			/*
				_index:   key=streamID, value=version
				_content: key=streamID, value=map[streamPos]globalPos
				_entries: key=globalPos, value=event
			*/
			_content := Content{streamID: stream.Name(), positions: make(map[uint64]uint64)}
			for idx, event := range stream.events {
				pos += 1
				_content.positions[idx] = pos
				entry, err := event.MarshalJSON()
				if err != nil {
					return err
				}
				if err = s.kvs.Put(persistance.KvsBucketEntries, strconv.FormatUint(pos, 10), entry); err != nil {
					return err
				}
			}
			content, err := _content.MarshalJSON()
			if err := s.kvs.Put(persistance.KvsBucketContent, stream.name, content); err != nil {
				return err
			}
			streamVersion := make([]byte, 8)
			binary.BigEndian.PutUint64(streamVersion, stream.version)
			if err := s.kvs.Put(persistance.KvsBucketIndex, stream.name, streamVersion); err != nil {
				return err
			}

		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *KVSService) read(_ context.Context, cmd ReadCommand) ([]Stream, error) {
	result := make([]Stream, 0)
	for _, stream := range cmd.streams {
		contentData, err := s.kvs.Get(persistance.KvsBucketContent, stream)
		if err != nil {
			return result, err
		}
		var content Content
		if err := content.UnmarshalJSON(contentData); err != nil {
			return result, err
		}
		streamData, err := s.kvs.Get(persistance.KvsBucketEntries, stream)
		if err != nil {
			return result, err
		}
		var stream Stream
		if err := stream.UnmarshalJSON(streamData); err != nil {
			return result, err
		}
		if uint64(len(elem.)) != elem.version {
			return result, fmt.Errorf("[ERROR]\t %T.read !!! version mismatch in stream <%s>: version=%d, events=%d",
				s, elem.name, elem.version, len(elem.events))
		}
		result = append(result, elem)
	}
	return result, nil
}

func (s *KVSService) readAt(_ context.Context, cmd ReadAtCommand) ([]Stream, error) {
	result := make([]Stream, 0)
	for _, stream := range cmd.streams {
		streamData, err := s.kvs.Get(persistance.KvsBucketContent, stream)
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

func (s *KVSService) firstAppend(streamData StreamData) *Stream {
	sort.Slice(streamData.events, func(i, j int) bool {
		return streamData.events[i].OccurredAt().Before(streamData.events[j].OccurredAt())
	})
	entries := getEntries(streamData)
	return buildStream(streamData.name, 0, entries)
}

func (s *KVSService) nextAppend(streamData StreamData) (*Stream, error) {
	raw, err := s.kvs.Get(persistance.KvsBucketContent, streamData.name)
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

func (s *KVSService) readLogIndex() (uint64, error) {
	value, err := s.kvs.Get(persistance.KvsBucketIndex, KeyGlobalPos)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(value), nil
}
