package database

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/juju/errors"
)

func camelCaseToUnderscore(s string) string {
	dest := []byte{}

	src := []byte(s)
	lastUpper := false
	for idx, b := range src {
		if b >= 'A' && b <= 'Z' {
			if !lastUpper || (idx < len(src)-1 && src[idx+1] >= 'a' && src[idx+1] <= 'z') {
				dest = append(dest, '_')
			}

			dest = append(dest, bytes.ToLower([]byte{b})[0])
			lastUpper = true
			continue
		}

		lastUpper = false
		dest = append(dest, b)
	}

	// Ignore the first underscore when building the final string
	if dest[0] == '_' {
		dest = dest[1:]
	}

	return string(dest)
}

type field struct {
	name string
	json bool
	gob  bool
}

func getSerializableFields(model reflect.Type) ([]*field, string, error) {
	fields := []*field{}
	columns := []string{}

	nfields := model.NumField()
	for i := 0; i < nfields; i++ {
		f := model.Field(i)
		tag := f.Tag.Get("db")

		// Ignore private fields
		if f.Name[0] >= 'a' && f.Name[0] <= 'z' {
			continue
		}

		// Ignore fields tagged with a dash
		if tag == "-" {
			continue
		}

		parts := strings.Split(tag, ",")

		fieldName := f.Name
		if len(parts) > 0 && parts[0] != "" {
			fieldName = parts[0]
		}

		var json, gob bool
		if len(parts) > 1 {
			switch parts[1] {
			case "serialize-json":
				json = true
			case "serialize-gob":
				gob = true
			case "":
				// Ignore empty parts
			default:
				return nil, "", errors.Errorf("unrecognized field tag: %s", parts[1])
			}
		}

		fields = append(fields, &field{
			name: fieldName,
			json: json,
			gob:  gob,
		})
		columns = append(columns, camelCaseToUnderscore(f.Name))
	}

	for i, column := range columns {
		columns[i] = fmt.Sprintf("`%s`", column)
	}

	return fields, strings.Join(columns, ", "), nil
}

func pluralize(singular string) string {
	if strings.HasSuffix(singular, "s") {
		return singular
	}
	if strings.HasSuffix(singular, "y") {
		return fmt.Sprintf("%sies", singular[:len(singular)-1])
	}

	return singular + "s"
}

func getPrimaryKeyFieldName(model interface{}) (string, error) {
	modelValue := reflect.ValueOf(model)
	modelType := reflect.TypeOf(model).Elem()

	// Some sanity checks about the model
	if modelValue.Kind() != reflect.Ptr || modelValue.Elem().Kind() != reflect.Struct {
		return "", errors.New("model should be a pointer to a struct")
	}

	var primary string

	// Try to find the primary field
	nfields := modelType.NumField()
	for i := 0; i < nfields; i++ {
		field := modelType.Field(i)

		// A field named ID is the default if no other tag is present
		if field.Name == "ID" && primary == "" {
			primary = "ID"
			continue
		}

		// A field tagged with `db:"primary-key"` would be marked as the primary key
		if field.Tag.Get("db") == "primary-key" {
			if primary != "" {
				return "", errors.New("model has several fields tagged as primary keys")
			}

			primary = field.Name
		}
	}

	if primary == "" {
		return "", errors.New("model does not have a recognizable primary key column")
	}

	return primary, nil
}

func getTableName(modelType reflect.Type) string {
	if _, ok := modelType.MethodByName("TableName"); ok {
		modelValue := reflect.New(modelType)
		return modelValue.MethodByName("TableName").Call([]reflect.Value{})[0].Interface().(string)
	}

	tableName := camelCaseToUnderscore(modelType.Name())
	tableName = pluralize(tableName)

	return tableName
}

func hasOperator(column string) bool {
	operators := []string{"=", "<>", "<", ">", "<=", ">=", "IN", "IS"}
	for _, op := range operators {
		if strings.Contains(column, op) {
			return true
		}
	}

	return false
}

func serializeField(field *field, modelValueElem reflect.Value) (interface{}, error) {
	rawValue := modelValueElem.FieldByName(field.name).Interface()

	if field.json {
		// Serialize with JSON when the field requires it
		buffer := bytes.NewBuffer(nil)
		if err := json.NewEncoder(buffer).Encode(rawValue); err != nil {
			return nil, errors.Trace(err)
		}

		return buffer.Bytes(), nil
	}

	if field.gob {
		// Serialize with gob when the field requires it
		buffer := bytes.NewBuffer(nil)
		if err := gob.NewEncoder(buffer).Encode(rawValue); err != nil {
			return nil, errors.Trace(err)
		}

		return buffer.Bytes(), nil
	}

	// Use the raw value if we're not a JSON-serializable field
	return rawValue, nil
}
