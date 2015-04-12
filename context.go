package database

import (
	"github.com/juju/errors"
	"golang.org/x/net/context"
)

type key int

var keyDatabase key

// WithContext opens a new connection to a remote database and adds it to the context
func WithContext(ctx context.Context, username, password, address, database string) (context.Context, error) {
	db, err := Connect(username, password, address, database)
	if err != nil {
		return ctx, errors.Trace(err)
	}

	return context.WithValue(ctx, keyDatabase, db), nil
}

// FromContext returns the database stored in the context
func FromContext(ctx context.Context) *Connection {
	return ctx.Value(keyDatabase).(*Connection)
}
