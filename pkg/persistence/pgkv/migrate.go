package pgkv

import (
	"database/sql"
	"log"
	"time"
)

type change struct {
	id       string
	commands []string
}

var changes = []change{
	{
		id: "2025-05-22-001_CREATE_BUCKETS",
		commands: []string{
			`CREATE TABLE BUCKETS (
				BUCKET_ID character varying(70) NOT NULL,
				BUCKET_KEY character varying(200) NOT NULL,
				BUCKET_VALUE bytea NOT NULL,
				CONSTRAINT PK_BUCKETS PRIMARY KEY (BUCKET_ID, BUCKET_KEY)
			);`,
		},
	},
}

// Migrate will prepare the database for proper use.
func (p *PostgresKVS) Migrate() error {
	if err := assertChangeRepository(p.db); err != nil {
		return err
	}
	for _, c := range changes {
		old, err := applyChange(p.db, c.id, c.commands)
		if err != nil {
			log.Printf("!! could not apply change %q - %s", c.id, err.Error())
			log.Printf("!!!! migration abort !!!!")
			return err
		}
		if old {
			log.Printf("-- change %q already loaded", c.id)
		} else {
			log.Printf("** change %q applied", c.id)
		}
	}
	log.Printf("---- migration complete ----")
	return nil
}

func assertChangeRepository(db *sql.DB) error {
	var migrationsCount string
	err := db.QueryRow("select count(*) from CHANGES").Scan(&migrationsCount)
	if err == nil {
		log.Printf("---- starting migration on a database of %s migrations", migrationsCount)
		return nil
	}
	query := "" +
		"create table CHANGES ( " +
		"  ID VARCHAR(128) not null unique, " +
		"  STAMP VARCHAR(35) not null " +
		")"
	_, err = db.Exec(query)
	if err != nil {
		log.Printf("could not create change repository - %s", err.Error())
		return err
	}
	return nil
}

func applyChange(db *sql.DB, id string, commands []string) (bool, error) {
	exists, err := changeExists(db, id)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	for _, command := range commands {
		_, err = tx.Exec(command)
		if err != nil {
			_ = tx.Rollback()
			return false, err
		}
	}
	_, err = tx.Exec("insert into CHANGES (ID, STAMP) values ($1, $2)", id, time.Now().Format(time.RFC3339))
	if err != nil {
		_ = tx.Rollback()
		return false, err
	}
	_ = tx.Commit()
	return false, err
}

func changeExists(db *sql.DB, id string) (bool, error) {
	rows, err := db.Query("select 1 from CHANGES where ID=$1", id)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}
