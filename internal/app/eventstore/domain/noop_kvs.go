package domain

import "github.com/openyard/eventstore/internal/app/persistance"

var (
	_ persistance.KeyValueStore  = (*noopKVS)(nil)
	_ persistance.KeyValueStoreX = (*noopKVSX)(nil)
)

type noopKVS struct{}

func (n noopKVS) Put(bucket, key string, value []byte) error {
	return nil
}

func (n noopKVS) Get(bucket, key string) ([]byte, error) {
	return make([]byte, 0), nil
}

func (n noopKVS) WithTx(fn ...func() error) error {
	for _, f := range fn {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

type noopKVSX struct{}

func (n noopKVSX) Put(bucket string, keys []string, values [][]byte) error {
	return nil
}

func (n noopKVSX) Get(bucket string, keys []string) (map[string][]byte, error) {
	return make(map[string][]byte), nil
}

func (n noopKVSX) WithTx(fn ...func() error) error {
	for _, f := range fn {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}
