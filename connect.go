package database

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/juju/errors"

	// It should be imported to register the MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

var (
	connectMutex sync.Mutex
	connections  = map[string]*Connection{}
)

// Connect opens a new connection to a remote database
func Connect(username, password, address, database string) (*Connection, error) {
	connectMutex.Lock()
	defer connectMutex.Unlock()

	dsn := fmt.Sprintf("%s:%s@%s/%s?charset=utf8&parseTime=true", username, password, address, database)

	if connections[dsn] == nil {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, errors.Trace(err)
		}

		// Fix an EOF problem when the connection sits idle for a long time and MySQL
		// closes the other side of the pipe unexpectly.
		// We're waiting for the golang team to fix the implementation of database/sql
		// before we can remove this hack:
		//   https://github.com/golang/go/issues/9851
		db.SetMaxIdleConns(0)

		// Test the connection is alive and working correctly
		if _, err := db.Exec("SELECT 1 = 1;"); err != nil {
			return nil, errors.Trace(err)
		}

		connections[dsn] = &Connection{DB: db}
	}

	return connections[dsn], nil
}
