package domain_test

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/openyard/eventstore/internal/app/eventstore/domain"
	"github.com/openyard/eventstore/internal/app/persistance"
	"github.com/openyard/eventstore/pkg/persistence/memkv"
	"github.com/stretchr/testify/assert"
)

var (
	ts, _  = time.Parse(time.RFC3339Nano, "2025-06-25T18:17:30.123456789Z")
	events = []*domain.Event{
		domain.NewEventAt("1", "test/v1.event", "4711", ts, nil),
		domain.NewEventAt("2", "test/v1.event", "4711", ts.Add(time.Second), nil),
		domain.NewEventAt("3", "test/v1.event", "4711", ts.Add(time.Second*2), nil),
		domain.NewEventAt("4", "test/v1.event", "4711", ts.Add(time.Second*3), nil),
		domain.NewEventAt("5", "test/v1.event", "4711", ts.Add(time.Second*4), nil),
	}
	streamData = domain.NewStreamData("test-stream", 0, events...)
)

func TestAppendCmd(t *testing.T) {
	kvs := memkv.NewMemoryKVS(persistance.KvsBucketIndex, persistance.KvsBucketContent, persistance.KvsBucketEntries)
	sut := domain.NewKVSService(domain.WithKeyValueStore(kvs))
	err := sut.HandleFunc(domain.NewCommand(context.Background(), domain.AppendCmd, domain.Append(streamData)))
	assert.NoError(t, err)
	streams, err := sut.QueryFunc(domain.NewCommand(context.Background(), domain.ReadCmd, domain.Read([]string{"test-stream"}...)))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(streams))
	testStream := streams[0]
	assert.Equal(t, "test-stream", testStream.Name())
	assert.Equal(t, uint64(5), testStream.Version())
	assert.Equal(t, "1", testStream.Events()[0].ID())
	assert.Equal(t, "2", testStream.Events()[1].ID())
	assert.Equal(t, "3", testStream.Events()[2].ID())
	assert.Equal(t, "4", testStream.Events()[3].ID())
	assert.Equal(t, "5", testStream.Events()[4].ID())

	ev1, err := kvs.Get(persistance.KvsBucketEntries, strconv.FormatUint(uint64(1), 10))
	assert.NoError(t, err)
	var entry1 domain.Entry
	err = json.Unmarshal(ev1, &entry1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), entry1.GlobalPos)
	assert.Equal(t, "test-stream", entry1.Stream)
	assert.Equal(t, uint64(1), entry1.StreamPos)
	assert.Equal(t, uint64(1), entry1.Event.ID())
	assert.Equal(t, "test/v1.event", entry1.Event.Name())
}
