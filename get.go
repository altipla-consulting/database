package database

import (
	"fmt"
	"reflect"

	"github.com/juju/errors"
	"golang.org/x/net/context"
)

// GetAll returns all the models a table has.
func GetAll(ctx context.Context, output interface{}) error {
	return NewQuery().GetAll(ctx, output)
}

// GetByID is a shortcut to querying the model by ID.
func GetByID(ctx context.Context, model interface{}, id interface{}) error {
	keyFieldName, err := getPrimaryKeyFieldName(model)
	if err != nil {
		return errors.Trace(err)
	}
	keyColumn := camelCaseToUnderscore(keyFieldName)

	query := NewQuery().Where(fmt.Sprintf("`%s` = ?", keyColumn), id)
	if err := query.Get(ctx, model); err != nil {
		return errors.Trace(err)
	}

	return nil
}

// GetIndexedByID fills a map[int64]*MyModel with the models fetched using
// the list of ids. It's an error if an id of the list is not present (ErrNoSuchEntity)
func GetIndexedByID(ctx context.Context, result interface{}, ids []int64) error {
	resultType := reflect.TypeOf(result)
	resultValue := reflect.ValueOf(result)

	// Some sanity checks about the result
	if resultType.Kind() != reflect.Ptr || resultType.Elem().Kind() != reflect.Map {
		return errors.New("result should be a pointer to a map")
	}
	resultElem := resultType.Elem()
	if resultElem.Key().Kind() != reflect.Int64 {
		return errors.New("keys of the map should be int64")
	}

	// Extract the key field name
	reflectElemValue := reflect.New(resultElem.Elem().Elem())
	getPKFieldName := reflect.ValueOf(getPrimaryKeyFieldName)
	ret := getPKFieldName.Call([]reflect.Value{reflectElemValue})
	keyFieldName, err := ret[0], ret[1]
	if !err.IsNil() {
		return errors.Trace(err.Interface().(error))
	}
	keyFieldNameStr := keyFieldName.Interface().(string)
	keyColumn := camelCaseToUnderscore(keyFieldNameStr)

	// Do nothing if there is no items to fetch
	if len(ids) == 0 {
		return nil
	}

	// Query the models
	query := NewQuery().Where(fmt.Sprintf("`%s` IN (?)", keyColumn), ids)
	getAll := reflect.ValueOf(query.GetAll)
	models := reflect.New(reflect.SliceOf(resultElem.Elem()))
	params := []reflect.Value{
		reflect.ValueOf(ctx),
		models,
	}
	if err := getAll.Call(params)[0]; !err.IsNil() {
		return errors.Trace(err.Interface().(error))
	}

	// Check that all IDs were present
	modelsElem := models.Elem()
	if modelsElem.Len() != len(ids) {
		return ErrNoSuchEntity
	}

	// Copy the models to the map
	resultValueElem := resultValue.Elem()
	for i := 0; i < modelsElem.Len(); i++ {
		item := modelsElem.Index(i)
		resultValueElem.SetMapIndex(item.Elem().FieldByName(keyFieldNameStr), item)
	}

	return nil
}
