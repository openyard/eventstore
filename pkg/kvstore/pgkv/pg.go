package pgkv

import (
	"database/sql"

	"github.com/openyard/eventstore/internal/app/kvstore"
)

var _ kvstore.KeyValueStore = (*PostgresKVS)(nil)

type PostgresKVS struct {
	db *sql.DB
}

func NewPostgresKVS(db *sql.DB, buckets ...string) *PostgresKVS {
	assertBuckets(db, buckets...)
	return &PostgresKVS{db: db}
}

func (p *PostgresKVS) AssertBucket(bucket string) error {
	//TODO implement me
	panic("implement me")
}

func (p *PostgresKVS) Put(bucket, key string, value []byte) error {
	//TODO implement me
	panic("implement me")
}

func (p *PostgresKVS) Get(bucket, key string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PostgresKVS) WithTx(fn ...func() error) error {
	panic("implement me")
	//TODO implement me
}

func (p *PostgresKVS) Rollback() {
	//TODO implement me
	panic("implement me")
}

func assertBuckets(db *sql.DB, buckets ...string) {

}
