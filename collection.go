package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Collection struct {
	sess *sql.DB
	conditions []Condition
	orders        []string
	offset, limit int64
	model Model
	props []*Property
}

func newCollection(db *Database, model Model) *Collection {
	props, err := extractModelProps(model)
	if err != nil {
		panic(err)
	}

	c := &Collection{
		sess: db.sess,
		model: model,
		props: props,
	}

	return c
}

func (c *Collection) Get(ctx context.Context, instance Model) error {
	b := newSQLBuilder(instance)
	for _, prop := range c.props {
		b.AddProperty(prop)

		if prop.PrimaryKey {
			b.Condition(&simpleCondition{prop.Name, prop.Value})
		}
	}

	statement, values := b.SelectSQL()
	if isDebug() {
		log.Println("database [Get]:", statement)
	}

	var pointers []interface{}
	for _, prop := range c.props {
		pointers = append(pointers, prop.Pointer)
	}
	if err := db.sess.QueryRowContext(ctx, statement, values...).Scan(pointers...); err != nil {
		if err == sql.ErrNoRows {
			return ErrNoSuchEntity
		}

		return err
	}

	b.Hydrate()

	if h, ok := instance.(ModelTrackingAfterGetHooker); ok {
		if err := h.ModelTrackingAfterGet(c.props); err != nil {
			return err
		}
	}

	return nil
}

func (c *Collection) Put(ctx context.Context, instance Model) error {
	b := newSQLBuilder(instance)

	var q string
	var values []interface{}
	if instance.IsInserted() {
		for _, prop := range c.props {
			if prop.PrimaryKey {
				b.Condition(&simpleCondition{prop.Name, prop.Value})
				continue
			}

			if prop.OmitEmpty && isZero(prop.Value) {
				continue
			}

			b.AddProperty(prop)
		}

		q, values = b.UpdateSQL()
	} else {
		for _, prop := range c.props {
			if prop.OmitEmpty && isZero(prop.Value) {
				continue
			}

			b.AddProperty(prop)
		}

		q, values = b.InsertSQL()
	}
	if isDebug() {
		log.Println("database [Put]:", q)
	}

	result, err := db.sess.ExecContext(ctx, q, values...)
	if err != nil {
		return err
	}

	var pks int
	for _, prop := range c.props {
		if prop.PrimaryKey {
			pks++
		}
	}
	if pks == 1 && !instance.IsInserted() {
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("database: cannot get last inserted id: %s", err)
		}

		v := reflect.ValueOf(instance).Elem()
		v.FieldByName(getPrimaryKeyField(c.props)).Set(reflect.ValueOf(id))
	}

	if h, ok := instance.(ModelTrackingAfterPutHooker); ok {
		if err := h.ModelTrackingAfterPut(c.props); err != nil {
			return err
		}
	}

	return nil
}

func (c *Collection) Filter(sql string, value interface{}) *Collection {
	return c.FilterCond(&simpleCondition{sql, value})
}

func (c *Collection) FilterCond(condition Condition) *Collection {
	c.conditions = append(c.conditions, condition)
	return c
}

func (c *Collection) Offset(offset int64) *Collection {
	c.offset = offset
	return c
}

func (c *Collection) Limit(limit int64) *Collection {
	c.limit = limit
	return c
}

func (c *Collection) Order(column string) *Collection {
	if strings.Contains(column, " ") {
		panic("call Order multiple times, do not pass multiple columns")
	}

	if strings.HasPrefix(column, "-") {
		column = fmt.Sprintf("%s DESC", column[1:])
	} else {
		column = fmt.Sprintf("%s ASC", column)
	}

	c.orders = append(c.orders, column)
	return c
}

func (c *Collection) Delete(ctx context.Context, instance Model) error {
	b := newSQLBuilder(instance)
	for _, prop := range c.props {
		if prop.PrimaryKey {
			b.Condition(&simpleCondition{prop.Name, prop.Value})
		}
	}

	statement, values := b.DeleteSQL()
	if isDebug() {
		log.Println("database [Delete]:", statement)
	}

	if _, err := db.sess.ExecContext(ctx, statement, values...); err != nil {
		return err
	}

	if h, ok := instance.(ModelTrackingAfterDeleteHooker); ok {
		if err := h.ModelTrackingAfterDelete(c.props); err != nil {
			return err
		}
	}

	return nil
}

func (c *Collection) Iterator(ctx context.Context) (*Iterator, error) {
	b := newSQLBuilder(c.model)

	sql, values := b.SelectSQL()
	if isDebug() {
		log.Println("database [Iterator]:", sql)
	}

	rows, err := c.sess.QueryContext(ctx, sql, values...)
	if err != nil {
		return nil, err
	}

	return &Iterator{rows, c.props}, nil
}

func (c *Collection) GetAll(ctx context.Context, models interface{}) error {
	v := reflect.ValueOf(models)
	t := reflect.TypeOf(models)

	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("database: pass a pointer to a slice to GetAll")
	}

	v = v.Elem()
	t = t.Elem()

	if v.Kind() != reflect.Slice {
		return fmt.Errorf("database: pass a slice to GetAll")
	}

	dest := reflect.MakeSlice(t, 0, 0)

	it, err := c.Iterator(ctx)
	if err != nil {
		return err
	}
	defer it.Close()

	for {
		model := reflect.New(t.Elem().Elem())
		if err := it.Next(model.Interface().(Model)); err != nil {
			if err == Done {
				break
			}

			return err
		}

		dest = reflect.Append(dest, model)
	}

	v.Set(dest)

	return nil
}


func (c *Collection) Count(ctx context.Context, model Model) (int64, error) {
	b := newSQLBuilder(c.model)

	sql, values := b.SelectSQLCols("COUNT(*)")
	if isDebug() {
		log.Println("database [Count]:", sql)
	}

	var n int64
	if err := db.sess.QueryRowContext(ctx, sql, values...).Scan(&n); err != nil {
		return 0, err
	}

	return n, nil
}
