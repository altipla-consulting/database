package database

import (
	"golang.org/x/net/context"
)

type key int

var keyDatabase key

// NewContext creates a new context containing the database
func NewContext(ctx context.Context, db *DB) context.Context {
	return context.WithValue(ctx, keyDatabase, db)
}

// FromContext returns the database stored in the context
func FromContext(ctx context.Context) *DB {
	return ctx.Value(keyDatabase).(*DB)
}
