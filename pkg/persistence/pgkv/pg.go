package pgkv

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/openyard/eventstore/internal/app/persistance"
)

var (
	_    persistance.KeyValueStore = (*PostgresKVS)(nil)
	zero                           = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
)

type PostgresKVS struct {
	db *sql.DB
}

func NewPostgresKVS(db *sql.DB) *PostgresKVS {
	p := &PostgresKVS{db: db}
	assertNoErr(p.Migrate())
	return p
}

func (p *PostgresKVS) Put(bucket, key string, value []byte) error {
	_, err := p.db.Exec(upsertBucketValue, bucket, key, value)
	assertNoErr(err)
	return nil
}

func (p *PostgresKVS) Get(bucket string, key string) ([]byte, error) {
	rows, err := p.db.Query(selectBucketValue, bucket, key)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	if !rows.Next() {
		return zero, fmt.Errorf("key <%s> not found in bucket <%s>", key, bucket)
	}
	var value []byte
	if err = rows.Scan(&value); err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, fmt.Errorf("key <%s> not unique in bucket <%s>", key, bucket)
	}
	return value, nil
}

func (p *PostgresKVS) WithTx(fn ...func() error) error {
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
	upsertBucketValue = `insert into BUCKETS (BUCKET_ID, BUCKET_KEY, BUCKET_VALUE) values ($1, $2, $3)
							on conflict on constraint PK_BUCKETS do 
							update set BUCKET_VALUE = EXCLUDED.BUCKET_VALUE`
	selectBucketValue = `select BUCKET_VALUE from BUCKETS where BUCKET_ID = $1 AND BUCKET_KEY = $2`
)

/*
	BUCKET_ID character varying(70) NOT NULL,
	BUCKET_KEY character varying(200) NOT NULL,
	BUCKET_VALUE bytea NOT NULL,
	CONTENT_TYPE character varying(200) NOT NULL,
	CONTENT_LENGTH numeric(16,0) NOT NULL,
	META jsonb DEFAULT NULL,
	STAMP character varying(35) NOT NULL,*/
