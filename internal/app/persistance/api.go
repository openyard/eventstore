package persistance

import "time"

const (
	KvsBucketIndex   = "_index"
	KvsBucketContent = "_content"
	KvsBucketEntries = "_entries"
)

// KeyValueStore provides an interface for a key-value-store
type KeyValueStore interface {
	Put(bucket, key string, value []byte) error
	Get(bucket, key string) ([]byte, error)
	WithTx(fn ...func() error) error
}

// KeyValueStoreX provides an interface for an experimental key-value-store able to process batches
type KeyValueStoreX interface {
	Put(bucket string, keys []string, values [][]byte) error
	Get(bucket string, keys []string) (map[string][]byte, error)
	WithTx(fn ...func() error) error
}

// SQLEventStore interface provides methods to read and write event-streams
type SQLEventStore interface {
	// Append adds the events to the assigned streams in one batch
	Append(changes ...Change) error
	// Delete removes streams from event-store physically
	Delete(streams ...string) error
	// Read loads all requested streams and returns them with its events
	Read(streams ...string) ([]Change, error)
	// ReadAt loads requested streams at a certain point in time and returns them with its events up to this point
	ReadAt(at time.Time, streams ...string) ([]Change, error)
}

// Change represents a single change incl. its expectedVersion in a stream.
// This struct is used to append multiple changes for multiple streams at once to an event-store
type Change struct {
	stream          string
	expectedVersion uint64
	aggregateID     string
	eventID         string
	correlationID   string // commandID or previousCorrelationID
	causationID     string // commandID or previousEventID
	eventName       string
	eventOccurredAt time.Time
	eventMeta       []byte // default: json, e.g. tracing information
	eventPayload    []byte // Content-Type known implicit through eventName
}
