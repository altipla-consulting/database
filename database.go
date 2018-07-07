package database

import (
	"database/sql"
	"fmt"
	"log"

	// Imports and registers the MySQL driver.
	_ "github.com/go-sql-driver/mysql"
)

// Database represents a reusable connection to a remote MySQL database.
type Database struct {
	sess *sql.DB
}

// Open starts a new connection to a remote MySQL database using the provided credentials
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

// Collection prepares a new collection using the table name of the model. It won't
// make any query, it only prepares the structs.
func (db *Database) Collection(model Model) *Collection {
	return newCollection(db, model)
}

// Close the connection. You should not use a database after closing it, nor any
// of its generated collections.
func (db *Database) Close() {
	db.sess.Close()
}

// Exec runs a raw SQL query in the database and returns nothing. It is
// recommended to use Collections instead.
func (db *Database) Exec(query string, params ...interface{}) error {
	_, err := db.sess.Exec(query, params...)
	return err
}

// QueryRow runs a raw SQL query in the database and returns the raw row from
// MySQL. It is recommended to use Collections instead.
func (db *Database) QueryRow(query string, params ...interface{}) *sql.Row {
	return db.sess.QueryRow(query, params...)
}
