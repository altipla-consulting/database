package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type Database struct {
	sess *sql.DB
}

func Open(ctx context.Context, credentials Credentials) (*Database, error) {
	if isDebug() {
		log.Println("database [Open]:", credentials)
	}

	sess, err := sql.Open("mysql", credentials.String())
	if err != nil {
		return nil, fmt.Errorf("database: cannot connect to mysql: %s", err)
	}

	sess.SetMaxOpenConns(3)
	sess.SetMaxIdleConns(0)

	if err := sess.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database: cannot ping mysql: %s", err)
	}

	return &Database{sess}, nil
}

func (db *Database) Collection(model Model) *Collection {
	return newCollection(db, model)
}

func (db *Database) Close() {
	db.sess.Close()
}

func (db *Database) Exec(ctx context.Context, query string, params ...interface{}) error {
	_, err := db.sess.ExecContext(ctx, query, params...)
	return err
}
