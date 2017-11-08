package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIteratorNextCallHooks(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	var models []*testingAutoModel
	require.Nil(t, testingsAuto.GetAll(ctx, &models))

	require.Len(t, models, 2)

	require.True(t, models[0].IsInserted())
	require.True(t, models[1].IsInserted())
}
