package database

import (
	"fmt"
	"reflect"

	"github.com/juju/errors"
	"golang.org/x/net/context"
)

// Delete removes a model from DB.
func Delete(ctx context.Context, model interface{}) error {
	modelValue := reflect.ValueOf(model)

	// Some sanity checks on the model
	if modelValue.Kind() != reflect.Ptr || modelValue.Elem().Kind() != reflect.Struct {
		return errors.New("model should be a pointer to a struct")
	}

	modelValueElem := modelValue.Elem()

	// Obtain the key field name, value and column name
	keyFieldName, err := getPrimaryKeyFieldName(model)
	if err != nil {
		return errors.Trace(err)
	}
	keyField := modelValueElem.FieldByName(keyFieldName)
	keyColumn := camelCaseToUnderscore(keyFieldName)
	keyFieldValue := modelValueElem.FieldByName(keyFieldName).Interface()

	// Check we have a primary key in the model
	if keyField.Interface() == reflect.Zero(keyField.Type()).Interface() {
		return errors.New("cannot delete a model without primary key")
	}

	// Remove the item
	return NewQuery().Where(fmt.Sprintf("`%s` = ?", keyColumn), keyFieldValue).Delete(ctx, model)
}
