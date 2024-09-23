package domain

import "github.com/openyard/eventstore/internal/app/kvstore"

var _ kvstore.KeyValueStore = (*noopKVs)(nil)

type noopKVs struct{}

func (n noopKVs) AssertBucket(_ string) error {
	return nil
}

func (n noopKVs) Put(bucket, key string, value []byte) error {
	return nil
}

func (n noopKVs) Get(bucket, key string) ([]byte, error) {
	return make([]byte, 0), nil
}

func (n noopKVs) WithTx(fn ...func() error) error {
	for _, f := range fn {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func (n noopKVs) Rollback() {
	// empty on purpose
}
