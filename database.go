package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type Database struct {
	sess *sql.DB
}

func Open(credentials Credentials) (*Database, error) {
	if isDebug() {
		log.Println("database [Open]:", credentials)
	}

	sess, err := sql.Open("mysql", credentials.String())
	if err != nil {
		return nil, fmt.Errorf("database: cannot connect to mysql: %s", err)
	}

	sess.SetMaxOpenConns(3)
	sess.SetMaxIdleConns(0)

	if err := sess.Ping(); err != nil {
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

func (db *Database) Exec(query string, params ...interface{}) error {
	_, err := db.sess.Exec(query, params...)
	return err
}

func (db *Database) QueryRow(query string, params ...interface{}) *sql.Row {
	return db.sess.QueryRow(query, params...)
}
