package domain

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/openyard/eventstore/internal/app/persistance"
	"log"
)

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

func (s *KVSXService) append(_ context.Context, cmd AppendCommand) error {
	if err := s.kvsx.WithTx(func() error {
		streams := make([]*Stream, 0)
		for _, streamData := range cmd.streamData {
			entries := getEntries(streamData)
			streams = append(streams, buildStream(streamData.name, uint64(len(entries)), entries))
		}
		streamsRaw, err := getStreamsRaw(streams)
		if err != nil {
			return err
		}
		bucketKeys := make([]string, 0, len(streamsRaw))
		bucketValues := make([][]byte, 0, len(streamsRaw))
		bucketIndices := make([][]byte, 0, len(streamsRaw))
		for i, stream := range streamsRaw {
			bucketKeys = append(bucketKeys, i)
			bucketValues = append(bucketValues, stream[0])
			bucketIndices = append(bucketIndices, stream[1])
		}
		if err = s.kvsx.Put(persistance.KvsBucketContent, bucketKeys, bucketValues); err != nil {
			return err
		}
		if err = s.kvsx.Put(persistance.KvsBucketIndex, bucketKeys, bucketIndices); err != nil {
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
		if err := elem.UnmarshalJSON(stream); err != nil {
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

func getStreamsRaw(streams []*Stream) (map[string][][]byte, error) {
	streamsRaw := make(map[string][][]byte)
	for _, stream := range streams {
		var raw []byte
		raw, err := stream.MarshalJSON()
		if err != nil {
			log.Printf("[ERROR]\t getStreamsRaw - marshaling error: %s", err)
			return nil, err
		}
		idx := make([]byte, 8)
		binary.BigEndian.PutUint64(idx, stream.version)
		streamsRaw[stream.name] = [][]byte{raw, idx}
	}
	return streamsRaw, nil
}
