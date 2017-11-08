package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	require.Nil(t, db.Exec(ctx, `INSERT INTO testing(code, name) VALUES ("foo", "foov"), ("bar", "barv")`))

	m := &testingModel{
		Code: "bar",
	}
	require.Nil(t, testings.Get(ctx, m))

	require.Equal(t, "barv", m.Name)
}

func TestGetNotFound(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "untouch",
	}
	require.EqualError(t, testings.Get(ctx, m), ErrNoSuchEntity.Error())
}

func TestGetNotTouchCols(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "untouched",
	}
	require.EqualError(t, testings.Get(ctx, m), ErrNoSuchEntity.Error())

	require.Equal(t, "untouched", m.Name)
}

func TestInsert(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "bar",
	}
	require.Nil(t, testings.Put(ctx, m))

	other := &testingModel{
		Code: "foo",
	}
	require.Nil(t, testings.Get(ctx, other))
	require.Equal(t, "bar", other.Name)
}

func TestInsertAuto(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingAutoModel{
		Name: "foo",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))
	require.EqualValues(t, m.ID, 1)

	m = &testingAutoModel{
		Name: "bar",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))
	require.EqualValues(t, m.ID, 2)

	other := &testingAutoModel{
		ID: 1,
	}
	require.Nil(t, testingsAuto.Get(ctx, other))
	require.Equal(t, "foo", other.Name)

	other = &testingAutoModel{
		ID: 2,
	}
	require.Nil(t, testingsAuto.Get(ctx, other))
	require.Equal(t, "bar", other.Name)
}

func TestUpdate(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "bar",
	}
	require.Nil(t, testings.Put(ctx, m))
	require.Nil(t, testings.Get(ctx, m))

	m.Name = "qux"
	require.Nil(t, testings.Put(ctx, m))

	other := &testingModel{
		Code: "foo",
	}
	require.Nil(t, testings.Get(ctx, other))
	require.Equal(t, "qux", other.Name)
}

func TestInsertAndUpdate(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "bar",
	}
	require.Nil(t, testings.Put(ctx, m))

	m.Name = "qux"
	require.Nil(t, testings.Put(ctx, m))

	other := &testingModel{
		Code: "foo",
	}
	require.Nil(t, testings.Get(ctx, other))
	require.Equal(t, "qux", other.Name)
}

func TestDelete(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "bar",
	}
	require.Nil(t, testings.Put(ctx, m))

	n, err := testings.Count(ctx, m)
	require.Nil(t, err)
	require.EqualValues(t, n, 1)

	require.Nil(t, testings.Delete(ctx, m))

	n, err = testings.Count(ctx, m)
	require.Nil(t, err)
	require.EqualValues(t, n, 0)
}
