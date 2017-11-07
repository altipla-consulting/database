package database

import (
	"fmt"
	"reflect"
	"strings"
)

type property struct {
	Name       string
	Value      interface{}
	PrimaryKey bool
	OmitEmpty  bool
	Pointer    interface{}
}

func extractModelProps(model Model) ([]*property, error) {
	v := reflect.ValueOf(model).Elem()
	t := reflect.TypeOf(model).Elem()

	props := []*property{}
	for i := 0; i < t.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)

		prop := &property{
			Name:    ft.Name,
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

		props = append(props, prop)
	}

	return props, nil
}
