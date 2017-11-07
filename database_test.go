package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

var db *Database

type testingModel struct {
	Code    string `db:"code,pk"`
	Name    string `db:"name"`
	Ignored string `db:"-"`
}

func (model *testingModel) TableName() string {
	return "testing"
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

	require.Equal(t, m.Name, "barv")
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

	require.Equal(t, m.Name, "untouched")
}
