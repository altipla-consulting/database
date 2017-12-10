package database

import (
	"errors"
	"strings"
)

var (
	// ErrNoSuchEntity is returned from a Get operation when there is not a model
	// that matches the query
	ErrNoSuchEntity = errors.New("database: no such entity")

	// Done is returned from the Next() method of an iterator when all results
	// have been read.
	Done = errors.New("query has no more results")

	// ErrConcurrentTransaction is returned when trying to update a model that has been
	// updated in the background by other process. This errors will prevent you from
	// potentially overwriting those changes.
	ErrConcurrentTransaction = errors.New("database: concurrent transaction")
)

type MultiError []error

func (merr MultiError) Error() string {
	var msg []string
	for _, err := range merr {
		if err == nil {
			msg = append(msg, "<nil>")
		} else {
			msg = append(msg, err.Error())
		}
	}

	return strings.Join(msg, "; ")
}

func (merr MultiError) HasError() bool {
	for _, err := range merr {
		if err != nil {
			return true
		}
	}

	return false
}
