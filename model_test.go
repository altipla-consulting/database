package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInsertedAfterGet(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()

	m := new(testingModel)
	tracking := m.Tracking()

	require.False(t, tracking.IsInserted())

	tracking.AfterGet(nil)

	require.True(t, tracking.IsInserted())
}

func TestInsertedAfterPut(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()

	m := new(testingModel)
	tracking := m.Tracking()

	require.False(t, tracking.IsInserted())

	tracking.AfterPut(nil)

	require.True(t, tracking.IsInserted())
}

func TestInsertedAfterDelete(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()

	m := new(testingModel)
	tracking := m.Tracking()
	tracking.AfterGet(nil)

	require.True(t, tracking.IsInserted())

	tracking.AfterDelete(nil)

	require.False(t, tracking.IsInserted())
}
