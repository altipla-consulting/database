package database

import (
	"fmt"
	"reflect"
	"strings"
)

var modelTrackingType = reflect.TypeOf(ModelTracking{})

type Model interface {
	TableName() string
	IsInserted() bool
}

type ModelTracking struct {
	inserted bool
}

func (tracking *ModelTracking) IsInserted() bool {
	return tracking.inserted
}

func (tracking *ModelTracking) ModelTrackingAfterGet(props []*Property) error {
	tracking.inserted = true
	return nil
}

func (tracking *ModelTracking) ModelTrackingAfterPut(props []*Property) error {
	tracking.inserted = true
	return nil
}

func (tracking *ModelTracking) ModelTrackingAfterDelete(props []*Property) error {
	tracking.inserted = false
	return nil
}

type Property struct {
	// Name of the column. Already escaped.
	Name string

	// Struct field name.
	Field string

	// Value of the field.
	Value interface{}

	// Pointer to the value of the field.
	Pointer interface{}

	// True if it's a primary key.
	PrimaryKey bool

	// Omit the column when the value is empty.
	OmitEmpty bool
}

func extractModelProps(model Model) ([]*Property, error) {
	v := reflect.ValueOf(model).Elem()
	t := reflect.TypeOf(model).Elem()

	props := []*Property{}
	for i := 0; i < t.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)

		if ft.Type == modelTrackingType {
			continue
		}

		prop := &Property{
			Name:    ft.Name,
			Field:   ft.Name,
			Value:   fv.Interface(),
			Pointer: fv.Addr().Interface(),
		}

		tag := ft.Tag.Get("db")
		if tag != "" {
			parts := strings.Split(tag, ",")
			if len(parts) > 2 {
				return nil, fmt.Errorf("database: unknown struct tag: %s", parts[1])
			}

			if parts[0] != "" {
				prop.Name = parts[0]
			}

			if len(parts) > 1 {
				switch parts[1] {
				case "pk":
					prop.PrimaryKey = true
					prop.OmitEmpty = true

				case "omitempty":
					prop.OmitEmpty = true

				default:
					return nil, fmt.Errorf("database: unknown struct tag: %s", parts[1])
				}
			}
		}

		if prop.Name == "-" {
			continue
		}

		// Escape the name inside the SQL query. It is NOT for security.
		prop.Name = fmt.Sprintf("`%s`", prop.Name)

		props = append(props, prop)
	}

	return props, nil
}

func getPrimaryKeyField(props []*Property) string {
	for _, prop := range props {
		if prop.PrimaryKey {
			return prop.Field
		}
	}

	return ""
}

func isZero(value interface{}) bool {
	switch v := value.(type) {
	case string:
		return len(v) == 0

	case int32:
		return v == 0

	case int64:
		return v == 0
	}

	return false
}
