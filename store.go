package database

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/juju/errors"
	"golang.org/x/net/context"
)

// CanUseSaveNotifier can be implemented in any model struct to notify the
// database package that the Save() function should not be used. This can be useful
// to avoid errors when the ID is not an autoincrement column but a custom one
// (we cannot tell the difference between INSERT and UPDATE with custom primary
// key columns).
type CanUseSaveNotifier interface {
	CanUseSave() bool
}

// Save stores a model in the DB. It will call Create() if there is no
// primary key on the model or Update() if it's a previously fetched struct.
// You can implement CanUseSaveNotifier to enforce the use of one of those two
// functions directly instead of this heuristic.
func Save(ctx context.Context, model interface{}) error {
	// Allow the models to forbid the use of save directly, because of non-empty
	// primary keys
	if notifier, ok := model.(CanUseSaveNotifier); ok {
		if !notifier.CanUseSave() {
			return errors.New("cannot use Save() with this kind of model")
		}
	}

	modelValue := reflect.ValueOf(model)

	// Some sanity checks about the model
	if modelValue.Kind() != reflect.Ptr || modelValue.Elem().Kind() != reflect.Struct {
		return errors.New("model should be a pointer to a struct")
	}

	// Get the key field of the model
	keyFieldName, err := getPrimaryKeyFieldName(modelValue.Interface())
	if err != nil {
		return errors.Trace(err)
	}
	keyField := modelValue.Elem().FieldByName(keyFieldName)
	isKeyEmpty := (keyField.Interface() == reflect.Zero(keyField.Type()).Interface())

	if isKeyEmpty {
		return errors.Trace(Create(ctx, model))
	}

	return errors.Trace(Update(ctx, model))
}

// Create stores a new model in the DB.
func Create(ctx context.Context, model interface{}) error {
	modelValue := reflect.ValueOf(model)
	modelType := reflect.TypeOf(model)

	// Some sanity checks about the model
	if modelValue.Kind() != reflect.Ptr || modelValue.Elem().Kind() != reflect.Struct {
		return errors.New("model should be a pointer to a struct")
	}

	modelValueElem := modelValue.Elem()
	modelTypeElem := modelType.Elem()

	// Build the table name
	tableName := getTableName(modelTypeElem)

	// Get the list of field names
	fields, columns, err := getSerializableFields(modelTypeElem)
	if err != nil {
		return errors.Trace(err)
	}

	var query string
	values := []interface{}{}

	// Prepare the placeholder question marks for the query
	placeholders := make([]string, len(fields))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	// Build the insert query
	query = fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", tableName, columns,
		strings.Join(placeholders, ", "))

	// Run before hooks
	if err := runBeforeSaveHook(modelValue); err != nil {
		return errors.Trace(err)
	}
	if err := runBeforeCreateHook(modelValue); err != nil {
		return errors.Trace(err)
	}

	// Add all the values to the query
	for _, field := range fields {
		serialized, err := serializeField(field, modelValueElem)
		if err != nil {
			return errors.Trace(err)
		}

		values = append(values, serialized)
	}

	// Exec the query
	conn := FromContext(ctx)
	result, err := conn.DB.Exec(query, values...)
	if err != nil {
		return errors.Annotate(err, query)
	}

	// Run after hooks
	if err := runAfterCreateHook(modelValue); err != nil {
		return errors.Trace(err)
	}
	if err := runAfterSaveHook(modelValue); err != nil {
		return errors.Trace(err)
	}

	// Store the last inserted id in the primary key field
	id, err := result.LastInsertId()
	if err != nil {
		return errors.Trace(err)
	}

	// If it's zero it's probably because the model does not have an autoincrement
	// value, or a non-integer primary key. Don't assign it if it occurs because it
	// will be an error.
	if id != 0 {
		keyFieldName, err := getPrimaryKeyFieldName(modelValue.Interface())
		if err != nil {
			return errors.Trace(err)
		}
		keyField := modelValueElem.FieldByName(keyFieldName)
		keyField.Set(reflect.ValueOf(id))
	}

	return nil
}

// Update stores the new data of an existing model in the DB.
func Update(ctx context.Context, model interface{}) error {
	_, err := UpdateRowsAffected(ctx, model)
	return errors.Trace(err)
}

// UpdateRowsAffected stores the new data of an existing model in the DB and returns
// the number of rows affected: one if it's successful or zero if you are using
// optimistic locking and the change failed.
func UpdateRowsAffected(ctx context.Context, model interface{}) (int64, error) {
	modelValue := reflect.ValueOf(model)
	modelType := reflect.TypeOf(model)

	// Some sanity checks about the model
	if modelValue.Kind() != reflect.Ptr || modelValue.Elem().Kind() != reflect.Struct {
		return 0, errors.New("model should be a pointer to a struct")
	}

	modelValueElem := modelValue.Elem()
	modelTypeElem := modelType.Elem()

	// Build the table name
	tableName := getTableName(modelTypeElem)

	// Get the list of field names
	fields, _, err := getSerializableFields(modelTypeElem)
	if err != nil {
		return 0, errors.Trace(err)
	}

	// Prepare the placeholder question marks for the query
	placeholders := make([]string, len(fields))
	for i, field := range fields {
		placeholders[i] = fmt.Sprintf("`%s` = ?", camelCaseToUnderscore(field.name))
	}

	// Run before hooks
	if err := runBeforeSaveHook(modelValue); err != nil {
		return 0, errors.Trace(err)
	}
	if err := runBeforeUpdateHook(modelValue); err != nil {
		return 0, errors.Trace(err)
	}

	// Add all the values to the query
	values := []interface{}{}
	for _, field := range fields {
		serialized, err := serializeField(field, modelValueElem)
		if err != nil {
			return 0, errors.Trace(err)
		}

		values = append(values, serialized)
	}

	// Add the primary key to filter the result we wanna update
	keyFieldName, err := getPrimaryKeyFieldName(modelValue.Interface())
	if err != nil {
		return 0, errors.Trace(err)
	}
	values = append(values, modelValueElem.FieldByName(keyFieldName).Interface())

	// Exec the query
	query := fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s` = ?", tableName,
		strings.Join(placeholders, ", "), camelCaseToUnderscore(keyFieldName))
	conn := FromContext(ctx)
	result, err := conn.DB.Exec(query, values...)
	if err != nil {
		return 0, errors.Trace(err)
	}

	// Run after hooks
	if err := runAfterUpdateHook(modelValue); err != nil {
		return 0, errors.Trace(err)
	}
	if err := runAfterSaveHook(modelValue); err != nil {
		return 0, errors.Trace(err)
	}

	// Number of rows affected, this is always present in the MySQL driver so it's
	// not a performance hit to call it always
	n, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Trace(err)
	}

	return n, nil
}
