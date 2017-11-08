package database

import (
  "database/sql"
  "reflect"
)

type Iterator struct {
  rows  *sql.Rows
  props []*Property
}

func (it *Iterator) Close() {
  it.rows.Close()
}

func (it *Iterator) Next(model Model) error {
  v := reflect.ValueOf(model).Elem()

  if err := it.rows.Err(); err != nil {
    return err
  }

  if !it.rows.Next() {
    if err := it.rows.Err(); err != nil {
      return err
    }

    it.Close()

    return Done
  }

  ptrs := make([]interface{}, len(it.props))
  for i, prop := range it.props {
    ptrs[i] = v.FieldByName(prop.Field).Addr().Interface()
  }
  if err := it.rows.Scan(ptrs...); err != nil {
    return err
  }

  for i, prop := range it.props {
    prop.Pointer = ptrs[i]
    prop.Value = reflect.ValueOf(prop.Pointer).Elem().Interface()
  }

  if h, ok := model.(ModelTrackingAfterGetHooker); ok {
    if err := h.ModelTrackingAfterGet(it.props); err != nil {
      return err
    }
  }

  return nil
}
