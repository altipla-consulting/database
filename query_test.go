package database

// import (
// 	"context"
// 	"testing"

// 	"github.com/stretchr/testify/require"
// )

// func TestGetAllEmpty(t *testing.T) {
// 	initDatabase(t)
// 	defer closeDatabase()
// 	ctx := context.Background()

// 	var models []*testingModel
// 	require.Nil(t, db.GetAll(ctx, &models))

// 	require.Len(t, models, 0)
// }

// func TestGetAll(t *testing.T) {
// 	initDatabase(t)
// 	defer closeDatabase()
// 	ctx := context.Background()

// 	m := &testingModel{
// 		Code: "foo",
// 		Name: "foo name",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	m = &testingModel{
// 		Code: "bar",
// 		Name: "bar name",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	var models []*testingModel
// 	require.Nil(t, db.GetAll(ctx, &models))

// 	require.Len(t, models, 2)

// 	require.Equal(t, "bar", models[0].Code)
// 	require.Equal(t, "bar name", models[0].Name)
// 	require.Equal(t, "foo", models[1].Code)
// 	require.Equal(t, "foo name", models[1].Name)
// }

// func TestGetAllFiltering(t *testing.T) {
// 	initDatabase(t)
// 	defer closeDatabase()
// 	ctx := context.Background()

// 	m := &testingModel{
// 		Code: "foo",
// 		Name: "test",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	m = &testingModel{
// 		Code: "bar",
// 		Name: "test",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	m = &testingModel{
// 		Code: "qux",
// 		Name: "ignore",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	var models []*testingModel
// 	require.Nil(t, db.Filter("name", "test").GetAll(ctx, &models))

// 	require.Len(t, models, 2)

// 	require.Equal(t, "bar", models[0].Code)
// 	require.Equal(t, "test", models[0].Name)
// 	require.Equal(t, "foo", models[1].Code)
// 	require.Equal(t, "test", models[1].Name)
// }

// func TestGetAllOperator(t *testing.T) {
// 	initDatabase(t)
// 	defer closeDatabase()
// 	ctx := context.Background()

// 	m := &testingAutoModel{
// 		Name: "foo",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	m = &testingAutoModel{
// 		Name: "bar",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	m = &testingAutoModel{
// 		Name: "ignore",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	var models []*testingAutoModel
// 	require.Nil(t, db.Filter("id <=", 2).GetAll(ctx, &models))

// 	require.Len(t, models, 2)

// 	require.Equal(t, "foo", models[0].Name)
// 	require.Equal(t, "bar", models[1].Name)
// }

// func TestGetAllOperatorAndPlaceholder(t *testing.T) {
// 	initDatabase(t)
// 	defer closeDatabase()
// 	ctx := context.Background()

// 	m := &testingAutoModel{
// 		Name: "foo",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	m = &testingAutoModel{
// 		Name: "bar",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	m = &testingAutoModel{
// 		Name: "baz",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	var models []*testingAutoModel
// 	require.Nil(t, db.Filter("name LIKE ?", "ba%").GetAll(ctx, &models))

// 	require.Len(t, models, 2)

// 	require.Equal(t, "bar", models[0].Name)
// 	require.Equal(t, "baz", models[1].Name)
// }

// func TestGetAllOperatorIN(t *testing.T) {
// 	initDatabase(t)
// 	defer closeDatabase()
// 	ctx := context.Background()

// 	m := &testingModel{
// 		Code: "foo",
// 		Name: "foo name",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	m = &testingModel{
// 		Code: "bar",
// 		Name: "bar name",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	m = &testingModel{
// 		Code: "qux",
// 		Name: "ignore",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	var models []*testingModel
// 	require.Nil(t, db.Filter("name IN", []string{"foo name", "bar name"}).GetAll(ctx, &models))

// 	require.Len(t, models, 2)

// 	require.Equal(t, "bar", models[0].Code)
// 	require.Equal(t, "foo", models[1].Code)
// }

// func TestGetOrder(t *testing.T) {
// 	initDatabase(t)
// 	defer closeDatabase()
// 	ctx := context.Background()

// 	m := &testingModel{
// 		Code: "foo",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	m = &testingModel{
// 		Code: "bar",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	var models []*testingModel
// 	require.Nil(t, db.Order("-code").GetAll(ctx, &models))

// 	require.Len(t, models, 2)

// 	require.Equal(t, "foo", models[0].Code)
// 	require.Equal(t, "bar", models[1].Code)
// }

// func TestMultipleFilters(t *testing.T) {
// 	initDatabase(t)
// 	defer closeDatabase()
// 	ctx := context.Background()

// 	m := &testingAutoModel{
// 		Name: "foo",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	m = &testingAutoModel{
// 		Name: "bar",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	m = &testingAutoModel{
// 		Name: "qux",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	var models []*testingAutoModel
// 	require.Nil(t, db.Filter("id >", 1).Filter("id <", 3).GetAll(ctx, &models))

// 	require.Len(t, models, 1)

// 	require.Equal(t, "bar", models[0].Name)
// }

// func TestCount(t *testing.T) {
// 	initDatabase(t)
// 	defer closeDatabase()
// 	ctx := context.Background()

// 	m := new(testingAutoModel)
// 	require.Nil(t, db.Put(ctx, m))

// 	m = new(testingAutoModel)
// 	require.Nil(t, db.Put(ctx, m))

// 	m = new(testingAutoModel)
// 	require.Nil(t, db.Put(ctx, m))

// 	n, err := db.Count(ctx, new(testingAutoModel))
// 	require.Nil(t, err)
// 	require.EqualValues(t, n, 3)
// }

// func TestCountFilter(t *testing.T) {
// 	initDatabase(t)
// 	defer closeDatabase()
// 	ctx := context.Background()

// 	m := new(testingAutoModel)
// 	require.Nil(t, db.Put(ctx, m))

// 	m = new(testingAutoModel)
// 	require.Nil(t, db.Put(ctx, m))

// 	m = new(testingAutoModel)
// 	require.Nil(t, db.Put(ctx, m))

// 	n, err := db.Filter("id >=", 2).Count(ctx, new(testingAutoModel))
// 	require.Nil(t, err)
// 	require.EqualValues(t, n, 2)
// }

// func TestCountLimit(t *testing.T) {
// 	initDatabase(t)
// 	defer closeDatabase()
// 	ctx := context.Background()

// 	m := &testingAutoModel{
// 		Name: "foo",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	m = &testingAutoModel{
// 		Name: "bar",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	m = &testingAutoModel{
// 		Name: "baz",
// 	}
// 	require.Nil(t, db.Put(ctx, m))

// 	var models []*testingAutoModel
// 	require.Nil(t, db.Limit(1).Offset(1).GetAll(ctx, &models))

// 	require.Len(t, models, 1)

// 	require.Equal(t, models[0].Name, "bar")
// }

// func TestNextCallHooks(t *testing.T) {
// 	initDatabase(t)
// 	defer closeDatabase()
// 	ctx := context.Background()

// 	m := new(testingAutoModel)
// 	require.Nil(t, db.Put(ctx, m))

// 	m = new(testingAutoModel)
// 	require.Nil(t, db.Put(ctx, m))

// 	var models []*testingAutoModel
// 	require.Nil(t, db.GetAll(ctx, &models))

// 	require.Len(t, models, 2)

// 	require.True(t, models[0].IsInserted())
// 	require.True(t, models[1].IsInserted())
// }
