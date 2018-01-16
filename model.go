package database

import (
	"fmt"
	"reflect"
	"strings"
)

var modelTrackingType = reflect.TypeOf(ModelTracking{})

type OnAfterPutHooker interface {
	OnAfterPutHook() error
}

type Model interface {
	TableName() string
	Tracking() *ModelTracking
}

type ModelTracking struct {
	Revision int64

	inserted bool
}

// Tracking returns the tracking instance of a model.
func (tracking *ModelTracking) Tracking() *ModelTracking {
	return tracking
}

// StoredRevision returns the stored revision when the model was retrieved.
// It can be -1 signalling that the model is a new one.
func (tracking *ModelTracking) StoredRevision() int64 {
	return tracking.Revision - 1
}

// IsInserted returns true if the model has been previously retrieved from the database.
func (tracking *ModelTracking) IsInserted() bool {
	return tracking.inserted
}

// AfterGet is a hook called after a model is retrieved from the database.
func (tracking *ModelTracking) AfterGet(props []*Property) error {
	tracking.inserted = true
	tracking.Revision++
	return nil
}

// AfterPut is a hook called after a model is updated or inserted in the database.
func (tracking *ModelTracking) AfterPut(props []*Property) error {
	tracking.inserted = true
	tracking.Revision++
	return nil
}

// AfterDelete is a hook called after a model is deleted from the database.
func (tracking *ModelTracking) AfterDelete(props []*Property) error {
	tracking.inserted = false
	return nil
}

// Property represents a field of the model mapped to a database column.
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
			props = append(props, &Property{
				Name:    "`revision`",
				Field:   "Revision",
				Value:   fv.FieldByName("Revision").Interface(),
				Pointer: fv.FieldByName("Revision").Addr().Interface(),
			})

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

func updatedProps(props []*Property, model Model) []*Property {
	v := reflect.ValueOf(model).Elem()

	var result []*Property
	for _, prop := range props {
		result = append(result, &Property{
			Name:       prop.Name,
			Field:      prop.Field,
			Value:      v.FieldByName(prop.Field).Interface(),
			Pointer:    v.FieldByName(prop.Field).Addr().Interface(),
			PrimaryKey: prop.PrimaryKey,
			OmitEmpty:  prop.OmitEmpty,
		})
	}

	return result
}
