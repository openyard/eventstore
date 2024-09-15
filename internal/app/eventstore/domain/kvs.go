package domain

const (
	KvsBucketIndex   = "_index"
	KvsBucketContent = "_content"
)

var _ KeyValueStore = (*noopKV)(nil)

// KeyValueStore provides an interface for a key-value-store
type KeyValueStore interface {
	AssertBucket(bucket string) error
	Put(bucket, key string, value []byte) error
	Get(bucket, key string) ([]byte, error)
	WithTx(fn ...func() error) error
	Rollback()
}

type noopKV struct{}

func (n noopKV) AssertBucket(_ string) error {
	return nil
}

func (n noopKV) Put(bucket, key string, value []byte) error {
	return nil
}

func (n noopKV) Get(bucket, key string) ([]byte, error) {
	return make([]byte, 0), nil
}

func (n noopKV) WithTx(fn ...func() error) error {
	for _, f := range fn {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func (n noopKV) Rollback() {
	// empty on purpose
}
