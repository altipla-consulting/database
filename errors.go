package database

import (
	"errors"
)

var (
	// ErrNoSuchEntity is returned from a Get operation when there is not a model
	// that matches the query
	ErrNoSuchEntity = errors.New("database: no such entity")

	// Done is returned from the Next() method of an iterator when all results
	// have been read.
	Done = errors.New("query has no more results")
)
