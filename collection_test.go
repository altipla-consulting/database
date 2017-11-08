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

	n, err := testings.Count(ctx)
	require.Nil(t, err)
	require.EqualValues(t, n, 1)

	require.Nil(t, testings.Delete(ctx, m))

	n, err = testings.Count(ctx)
	require.Nil(t, err)
	require.EqualValues(t, n, 0)
}

func TestGetAllEmpty(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	var models []*testingModel
	require.Nil(t, testings.GetAll(ctx, &models))

	require.Len(t, models, 0)
}

func TestGetAll(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "foo name",
	}
	require.Nil(t, testings.Put(ctx, m))

	m = &testingModel{
		Code: "bar",
		Name: "bar name",
	}
	require.Nil(t, testings.Put(ctx, m))

	var models []*testingModel
	require.Nil(t, testings.GetAll(ctx, &models))

	require.Len(t, models, 2)

	require.Equal(t, "bar", models[0].Code)
	require.Equal(t, "bar name", models[0].Name)
	require.Equal(t, "foo", models[1].Code)
	require.Equal(t, "foo name", models[1].Name)
}

func TestGetAllFiltering(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "test",
	}
	require.Nil(t, testings.Put(ctx, m))

	m = &testingModel{
		Code: "bar",
		Name: "test",
	}
	require.Nil(t, testings.Put(ctx, m))

	m = &testingModel{
		Code: "qux",
		Name: "ignore",
	}
	require.Nil(t, testings.Put(ctx, m))

	var models []*testingModel
	require.Nil(t, testings.Filter("name", "test").GetAll(ctx, &models))

	require.Len(t, models, 2)

	require.Equal(t, "bar", models[0].Code)
	require.Equal(t, "test", models[0].Name)
	require.Equal(t, "foo", models[1].Code)
	require.Equal(t, "test", models[1].Name)
}

func TestGetAllOperator(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingAutoModel{
		Name: "foo",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "bar",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "ignore",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	var models []*testingAutoModel
	require.Nil(t, testingsAuto.Filter("id <=", 2).GetAll(ctx, &models))

	require.Len(t, models, 2)

	require.Equal(t, "foo", models[0].Name)
	require.Equal(t, "bar", models[1].Name)
}

func TestGetAllOperatorAndPlaceholder(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingAutoModel{
		Name: "foo",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "bar",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "baz",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	var models []*testingAutoModel
	require.Nil(t, testingsAuto.Filter("name LIKE ?", "ba%").GetAll(ctx, &models))

	require.Len(t, models, 2)

	require.Equal(t, "bar", models[0].Name)
	require.Equal(t, "baz", models[1].Name)
}

func TestGetAllOperatorIN(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "foo name",
	}
	require.Nil(t, testings.Put(ctx, m))

	m = &testingModel{
		Code: "bar",
		Name: "bar name",
	}
	require.Nil(t, testings.Put(ctx, m))

	m = &testingModel{
		Code: "qux",
		Name: "ignore",
	}
	require.Nil(t, testings.Put(ctx, m))

	var models []*testingModel
	require.Nil(t, testings.Filter("name IN", []string{"foo name", "bar name"}).GetAll(ctx, &models))

	require.Len(t, models, 2)

	require.Equal(t, "bar", models[0].Code)
	require.Equal(t, "foo", models[1].Code)
}

func TestGetOrder(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
	}
	require.Nil(t, testings.Put(ctx, m))

	m = &testingModel{
		Code: "bar",
	}
	require.Nil(t, testings.Put(ctx, m))

	var models []*testingModel
	require.Nil(t, testings.Order("-code").GetAll(ctx, &models))

	require.Len(t, models, 2)

	require.Equal(t, "foo", models[0].Code)
	require.Equal(t, "bar", models[1].Code)
}

func TestMultipleFilters(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingAutoModel{
		Name: "foo",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "bar",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "qux",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	var models []*testingAutoModel
	require.Nil(t, testingsAuto.Filter("id >", 1).Filter("id <", 3).GetAll(ctx, &models))

	require.Len(t, models, 1)

	require.Equal(t, "bar", models[0].Name)
}

func TestCount(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	n, err := testingsAuto.Count(ctx)
	require.Nil(t, err)
	require.EqualValues(t, n, 3)
}

func TestCountFilter(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	n, err := testingsAuto.Filter("id >=", 2).Count(ctx)
	require.Nil(t, err)
	require.EqualValues(t, n, 2)
}

func TestLimit(t *testing.T) {
	initDatabase(t)
	defer closeDatabase()
	ctx := context.Background()

	m := &testingAutoModel{
		Name: "foo",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "bar",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "baz",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	var models []*testingAutoModel
	require.Nil(t, testingsAuto.Limit(1).Offset(1).GetAll(ctx, &models))

	require.Len(t, models, 1)

	require.Equal(t, models[0].Name, "bar")
}
