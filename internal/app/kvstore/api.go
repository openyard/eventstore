package kvstore

const (
	KvsBucketIndex   = "_index"
	KvsBucketContent = "_content"
)

// KeyValueStore provides an interface for a key-value-store
type KeyValueStore interface {
	AssertBucket(bucket string) error
	Put(bucket, key string, value []byte) error
	Get(bucket, key string) ([]byte, error)
	WithTx(fn ...func() error) error
	Rollback()
}
