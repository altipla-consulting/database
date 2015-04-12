package database

import (
	"database/sql"
)

// Connection wraps a raw connection to a database.
type Connection struct {
	DB    *sql.DB
	Debug bool
}
