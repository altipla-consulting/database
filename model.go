package database

import (
	"reflect"
	"fmt"
	"strings"
)

type Property struct {
	Name string
	Value interface{}
	PrimaryKey bool
	OmitEmpty bool
}

func extractModelProps(model Model) ([]*Property, error) {
	v := reflect.ValueOf(model)
	t := reflect.TypeOf(model)

	props := []*Property{}
	for i := 0; i < t.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)

		prop := &Property{
			Name: ft.Name,
			Value: fv.Interface(),
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

		props = append(props, prop)
	}

	return props, nil
}
