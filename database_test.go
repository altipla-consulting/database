package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

var db *Database

type testingModel struct {
	ModelTracking

	Code    string `db:"code,pk"`
	Name    string `db:"name"`
	Ignored string `db:"-"`
}

func (model *testingModel) TableName() string {
	return "testing"
}

type testingAutoModel struct {
	ModelTracking

	ID      int64  `db:"id,pk"`
	Name    string `db:"name"`
	Ignored string `db:"-"`
}

func (model *testingAutoModel) TableName() string {
	return "testing_auto"
}

func initDatabase(t *testing.T) {
	ctx := context.Background()

	var err error
	db, err = Open(ctx, Credentials{
		User:      "dev-user",
		Password:  "dev-password",
		Address:   "localhost",
		Database:  "test",
		Charset:   "utf8mb4",
		Collation: "utf8mb4_bin",
	})
	require.Nil(t, err)

	require.Nil(t, db.Exec(ctx, `DROP TABLE IF EXISTS testing`))
	err = db.Exec(ctx, `
    CREATE TABLE testing (
      code VARCHAR(191),
      name VARCHAR(191),

      PRIMARY KEY(code)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	require.Nil(t, db.Exec(ctx, `DROP TABLE IF EXISTS testing_auto`))
	err = db.Exec(ctx, `
    CREATE TABLE testing_auto (
      id INT(11) NOT NULL AUTO_INCREMENT,
      name VARCHAR(191),

      PRIMARY KEY(id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)
}

func closeDatabase() {
	db.Close()
}

func TestGet(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	require.Nil(t, db.Exec(ctx, `INSERT INTO testing(code, name) VALUES ("foo", "foov"), ("bar", "barv")`))

	m := &testingModel{
		Code: "bar",
	}
	require.Nil(t, db.Get(ctx, m))

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
	require.EqualError(t, db.Get(ctx, m), ErrNoSuchEntity.Error())
}

func TestGetNotTouchCols(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "untouched",
	}
	require.EqualError(t, db.Get(ctx, m), ErrNoSuchEntity.Error())

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
	require.Nil(t, db.Put(ctx, m))

	other := &testingModel{
		Code: "foo",
	}
	require.Nil(t, db.Get(ctx, other))
	require.Equal(t, "bar", other.Name)
}

func TestInsertAuto(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingAutoModel{
		Name: "foo",
	}
	require.Nil(t, db.Put(ctx, m))
	require.EqualValues(t, m.ID, 1)

	m = &testingAutoModel{
		Name: "bar",
	}
	require.Nil(t, db.Put(ctx, m))
	require.EqualValues(t, m.ID, 2)

	other := &testingAutoModel{
		ID: 1,
	}
	require.Nil(t, db.Get(ctx, other))
	require.Equal(t, "foo", other.Name)

	other = &testingAutoModel{
		ID: 2,
	}
	require.Nil(t, db.Get(ctx, other))
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
	require.Nil(t, db.Put(ctx, m))
	require.Nil(t, db.Get(ctx, m))

	m.Name = "qux"
	require.Nil(t, db.Put(ctx, m))

	other := &testingModel{
		Code: "foo",
	}
	require.Nil(t, db.Get(ctx, other))
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
	require.Nil(t, db.Put(ctx, m))

	m.Name = "qux"
	require.Nil(t, db.Put(ctx, m))

	other := &testingModel{
		Code: "foo",
	}
	require.Nil(t, db.Get(ctx, other))
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
	require.Nil(t, db.Put(ctx, m))

	n, err := db.Count(ctx, m)
	require.Nil(t, err)
	require.EqualValues(t, n, 1)

	require.Nil(t, db.Delete(ctx, m))

	n, err = db.Count(ctx, m)
	require.Nil(t, err)
	require.EqualValues(t, n, 0)
}
