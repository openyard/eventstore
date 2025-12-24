package domain

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"strconv"

	"github.com/openyard/eventstore/internal/app/persistance"
)

const KeyGlobalPos = "_globalPos"

var _ Service = (*KVSXService)(nil)

type KVSXServiceOpts func(*KVSXService)

type KVSXService struct {
	kvsx persistance.KeyValueStoreX
}

func NewKVSXService(opts ...KVSXServiceOpts) *KVSXService {
	s := &KVSXService{
		kvsx: &noopKVSX{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithKeyValueStoreX(kvsx persistance.KeyValueStoreX) KVSXServiceOpts {
	return func(s *KVSXService) {
		s.kvsx = kvsx
	}
}

func (s *KVSXService) HandleFunc(cmd Command) error {
	switch cmd.kind {
	case AppendCmd:
		return s.append(cmd.ctx, cmd.payload.(AppendCommand))
	default:
		return fmt.Errorf("unknown command: <%v>", cmd.kind)
	}
}

func (s *KVSXService) QueryFunc(cmd Command) ([]Stream, error) {
	switch cmd.kind {
	case ReadCmd:
		return s.read(cmd.ctx, cmd.payload.(ReadCommand))
	case ReadAtCmd:
		return s.readAt(cmd.ctx, cmd.payload.(ReadAtCommand))
	default:
		return nil, fmt.Errorf("unknown command: <%v>", cmd.kind)
	}
}

func (s *KVSXService) SubscribeFunc(cmd Command) ([]Entry, error) {
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

func (s *KVSXService) append(_ context.Context, cmd AppendCommand) error {
	if err := s.kvsx.WithTx(func() error {
		esIdx, err := s.readLogIndex()
		if err != nil {
			return err
		}
		streams := make([]*Stream, 0)
		for _, streamData := range cmd.streamData {
			entries := getEntries(streamData)
			streams = append(streams, buildStream(streamData.name, uint64(len(entries)), entries))
		}
		/*
			_index:   key=streamID, value=version
			_content: key=streamID, value=map[streamPos]globalPos
			_entries: key=globalPos, value=event
		*/
		streamIndices, err := getStreamsIndices(streams)
		if err != nil {
			return err
		}
		bucketKeys := make([]string, 0, len(streamIndices))
		bucketValues := make([][]byte, 0, len(streamIndices))
		for streamName, stream := range streamIndices {
			bucketKeys = append(bucketKeys, streamName)
			bucketValues = append(bucketValues, stream)
		}

		entriesRaw, err := getEntriesRaw(esIdx, streams)
		if err != nil {
			return err
		}
		contentKeys := make([]string, 0, len(entriesRaw))
		contentValues := make([][]byte, 0, len(entriesRaw))

		entryKeys := make([]string, 0)
		entryValues := make([][]byte, 0)
		for streamName, entries := range entriesRaw {
			contentKeys = append(contentKeys, streamName)
			contentValues = append(contentValues)
			for pos, entry := range entries {
				entryKeys = append(entryKeys, strconv.FormatUint(pos, 10))
				entryValues = append(entryValues, entry)
			}
		}

		if err = s.kvsx.Put(persistance.KvsBucketEntries, entryKeys, entryValues); err != nil {
			return err
		}
		if err = s.kvsx.Put(persistance.KvsBucketContent, contentKeys, contentValues); err != nil {
			return err
		}
		if err = s.kvsx.Put(persistance.KvsBucketIndex, bucketKeys, bucketValues); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *KVSXService) read(_ context.Context, cmd ReadCommand) ([]Stream, error) {
	result := make([]Stream, 0)
	streamData, err := s.kvsx.Get(persistance.KvsBucketContent, cmd.streams)
	if err != nil {
		return result, err
	}
	for _, stream := range streamData {
		var elem Stream
		if err = elem.UnmarshalJSON(stream); err != nil {
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

func (s *KVSXService) readAt(_ context.Context, cmd ReadAtCommand) ([]Stream, error) {
	result := make([]Stream, 0)
	streamData, err := s.kvsx.Get(persistance.KvsBucketContent, cmd.streams)
	if err != nil {
		return result, err
	}
	for _, stream := range streamData {
		var elem Stream
		if err := elem.UnmarshalJSON(stream); err != nil {
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

func (s *KVSXService) readLogIndex() (uint64, error) {
	values, err := s.kvsx.Get(persistance.KvsBucketIndex, []string{KeyGlobalPos})
	if err != nil {
		return 0, err
	}
	if len(values) != 1 {
		return 0, fmt.Errorf("[%4d] index error: expected <1> got <%d>", ErrLogIndex, len(values))
	}
	return binary.BigEndian.Uint64(values[KeyGlobalPos]), nil
}

func getStreamsIndices(streams []*Stream) (map[string][]byte, error) {
	streamsRaw := make(map[string][]byte)
	for _, stream := range streams {
		idx := make([]byte, 8)
		binary.BigEndian.PutUint64(idx, stream.version)
		streamsRaw[stream.name] = idx
	}
	return streamsRaw, nil
}

func getEntriesRaw(esIdx uint64, streams []*Stream) (map[string]map[uint64][]byte, error) {
	entriesRaw := make(map[string]map[uint64][]byte)
	for _, stream := range streams {
		entriesRaw[stream.name] = make(map[uint64][]byte, len(stream.events))
		for i, event := range stream.events {
			var raw []byte
			raw, err := event.MarshalJSON()
			if err != nil {
				log.Printf("[ERROR] getEntriesRaw - marshaling error: %s", err)
				return nil, err
			}
			entriesRaw[stream.name][i] = raw
		}
	}
	return entriesRaw, nil
}
