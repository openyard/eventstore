package memkv

import (
	"fmt"
	"log"
	"sync"

	"github.com/openyard/eventstore/internal/app/kvstore"
)

var (
	_    kvstore.KeyValueStore = (*MemoryKVS)(nil)
	zero                       = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
)

type MemoryKVS struct {
	sync.RWMutex
	buckets map[string]map[string][]byte // bucket, key, value
}

func NewMemoryKVS(buckets ...string) *MemoryKVS {
	memKVS := &MemoryKVS{buckets: make(map[string]map[string][]byte)}
	memKVS.assertBuckets(buckets...)
	return memKVS
}

func (m *MemoryKVS) AssertBucket(bucket string) error {
	if _, ok := m.buckets[bucket]; !ok {
		return fmt.Errorf("bucket (%s) not found", bucket)
	}
	return nil
}

func (m *MemoryKVS) Put(bucket, key string, value []byte) error {
	m.Lock()
	defer m.Unlock()
	if err := m.AssertBucket(bucket); err != nil {
		return err
	}
	m.buckets[bucket][key] = value
	return nil
}

func (m *MemoryKVS) Get(bucket, key string) ([]byte, error) {
	m.Lock()
	defer m.Unlock()
	if err := m.AssertBucket(bucket); err != nil {
		return zero, err
	}
	if value, ok := m.buckets[bucket][key]; ok {
		return value, nil
	}
	return zero, fmt.Errorf("key (%s:%s) not found", bucket, key)
}

func (m *MemoryKVS) Rollback() {
	// empty on purpose
}

func (m *MemoryKVS) WithTx(fn ...func() error) error {
	for _, f := range fn {

		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func (m *MemoryKVS) assertBuckets(IDs ...string) {
	defer log.Printf("assert buckets: %+v", IDs)
	for _, ID := range IDs {
		m.buckets[ID] = make(map[string][]byte)
	}
}
