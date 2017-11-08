package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testDB       *Database
	testings     *Collection
	testingsAuto *Collection
)

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
	testDB, err = Open(ctx, Credentials{
		User:      "dev-user",
		Password:  "dev-password",
		Address:   "localhost",
		Database:  "test",
		Charset:   "utf8mb4",
		Collation: "utf8mb4_bin",
	})
	require.Nil(t, err)

	require.Nil(t, testDB.Exec(ctx, `DROP TABLE IF EXISTS testing`))
	err = testDB.Exec(ctx, `
    CREATE TABLE testing (
      code VARCHAR(191),
      name VARCHAR(191),

      PRIMARY KEY(code)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	require.Nil(t, testDB.Exec(ctx, `DROP TABLE IF EXISTS testing_auto`))
	err = testDB.Exec(ctx, `
    CREATE TABLE testing_auto (
      id INT(11) NOT NULL AUTO_INCREMENT,
      name VARCHAR(191),

      PRIMARY KEY(id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
  `)
	require.Nil(t, err)

	testings = testDB.Collection(new(testingModel))
	testingsAuto = testDB.Collection(new(testingAutoModel))
}

func closeDatabase() {
	testDB.Close()
}
