package database

import (
	"github.com/juju/errors"
)

var (
	// ErrNoSuchEntity is returned from a Get operation when there is not a model
	// that matches the query
	ErrNoSuchEntity = errors.New("no such entity found in database")
)
