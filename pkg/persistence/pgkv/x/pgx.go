package x

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/lib/pq"
	"github.com/openyard/eventstore/internal/app/persistance"
)

var (
	_     persistance.KeyValueStoreX = (*PostgresKVSX)(nil)
	empty                            = make(map[string][]byte)
)

type PostgresKVSX struct {
	db *sql.DB
}

func NewPostgresKVSX(db *sql.DB, _ ...string) *PostgresKVSX {
	p := &PostgresKVSX{db: db}
	assertNoErr(p.Migrate())
	return p
}

func (p *PostgresKVSX) Put(bucket string, keys []string, values [][]byte) error {
	log.Println(bucket)
	log.Println(len(keys))
	log.Println(len(values))
	for i, v := range keys {
		log.Println(bucket, v, len(values[i]))
	}
	_, err := p.db.Exec(upsertBucketValues, bucket, pq.Array(keys), pq.Array(values))
	assertNoErr(err)
	return nil
}

func (p *PostgresKVSX) Get(bucket string, keys []string) (map[string][]byte, error) {
	rows, err := p.db.Query(selectBucketValues, bucket, keys)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	if !rows.Next() {
		return empty, fmt.Errorf("key <%s> not found in bucket <%s>", keys, bucket)
	}
	buckets := make(map[string][]byte)
	for rows.Next() {
		var bucketKey string
		var bucketValue []byte
		if err = rows.Scan(&bucketKey, &bucketValue); err != nil {
			return nil, err
		}
		buckets[bucketKey] = bucketValue
	}
	return buckets, nil
}

func (p *PostgresKVSX) WithTx(fn ...func() error) error {
	tx, err := p.db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	if err != nil {
		return err
	}
	for _, f := range fn {
		if err = f(); err != nil {
			log.Printf("[ERROR][%T.WithTx] %s", p, err)
			return errors.Join(err, tx.Rollback())
		}
	}
	return tx.Commit()
}

func assertNoErr(err error) {
	if err != nil {
		panic(err)
	}
}

const (
	upsertBucketValues = `insert into BUCKETS (BUCKET_ID, BUCKET_KEY, BUCKET_VALUE) select $1,* from unnest($2::text[], $3::bytea[])
							on conflict on constraint PK_BUCKETS do
							update set BUCKET_VALUE = EXCLUDED.BUCKET_VALUE`
	selectBucketValues = `select BUCKET_KEY, BUCKET_VALUE from BUCKETS where BUCKET_ID = $1 AND BUCKET_KEY = unnest($2::text[])`
)
