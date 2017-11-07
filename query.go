package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
)

type Query struct {
	sess          *sql.DB
	conditions    []Condition
	orders        []string
	offset, limit int64
}

type Condition interface {
	SQL() string
	Values() []interface{}
}

type simpleCondition struct {
	sql   string
	value interface{}
}

func (cond *simpleCondition) SQL() string {
	if !strings.Contains(cond.sql, " ") {
		return fmt.Sprintf("%s = ?", cond.sql)
	}

	if strings.Contains(cond.sql, " IN") {
		v := reflect.ValueOf(cond.value)
		placeholders := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			placeholders[i] = "?"
		}
		return fmt.Sprintf("%s (%s)", cond.sql, strings.Join(placeholders, ", "))
	}

	if !strings.Contains(cond.sql, "?") {
		return fmt.Sprintf("%s ?", cond.sql)
	}

	return cond.sql
}

func (cond *simpleCondition) Values() []interface{} {
	if strings.Contains(cond.sql, " IN") {
		v := reflect.ValueOf(cond.value)
		var values []interface{}
		for i := 0; i < v.Len(); i++ {
			values = append(values, v.Index(i).Interface())
		}
		return values
	}

	return []interface{}{cond.value}
}

func newQuery(db *Database, condition Condition) *Query {
	return &Query{
		sess:       db.sess,
		conditions: []Condition{condition},
	}
}

func newEmptyQuery(db *Database) *Query {
	return &Query{
		sess: db.sess,
	}
}

func (q *Query) Filter(sql string, value interface{}) *Query {
	return q.FilterCond(&simpleCondition{sql, value})
}

func (q *Query) FilterCond(condition Condition) *Query {
	q.conditions = append(q.conditions, condition)
	return q
}

func (q *Query) Order(column string) *Query {
	if strings.Contains(column, " ") {
		panic("call Order multiple times, do not pass multiple columns")
	}

	if strings.HasPrefix(column, "-") {
		column = fmt.Sprintf("%s DESC", column[1:])
	} else {
		column = fmt.Sprintf("%s ASC", column)
	}

	q.orders = append(q.orders, column)
	return q
}

func (q *Query) GetAll(ctx context.Context, models interface{}) error {
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

	it, err := q.Iterator(ctx, reflect.New(t.Elem().Elem()).Interface().(Model))
	if err != nil {
		return err
	}
	defer it.Close()

	for {
		model := reflect.New(t.Elem().Elem())
		if err := it.Next(model.Interface()); err != nil {
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

func (q *Query) Count(ctx context.Context, model Model) (int64, error) {
	var conds []string
	var values []interface{}
	for _, cond := range q.conditions {
		conds = append(conds, cond.SQL())
		values = append(values, cond.Values()...)
	}

	sqlq := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, model.TableName())
	if len(conds) > 0 {
		sqlq = fmt.Sprintf("%s WHERE %s", sqlq, strings.Join(conds, " AND "))
	}
	if len(q.orders) > 0 {
		sqlq = fmt.Sprintf("%s ORDER BY %s", sqlq, strings.Join(q.orders, ", "))
	}
	if isDebug() {
		log.Println("database [Count]:", sqlq)
	}

	var n int64
	if err := db.sess.QueryRowContext(ctx, sqlq, values...).Scan(&n); err != nil {
		return 0, err
	}

	return n, nil
}

func (q *Query) RandomRow(ctx context.Context, model Model) error {
	c, err := q.Count(ctx, model)
	if err != nil {
		return err
	}

	_ = c

	return nil
}

func (q *Query) Offset(offset int64) *Query {
	q.offset = offset
	return q
}

func (q *Query) Limit(limit int64) *Query {
	q.limit = limit
	return q
}

type Iterator struct {
	rows  *sql.Rows
	props []*Property
}

func (it *Iterator) Close() {
	it.rows.Close()
}

func (it *Iterator) Next(model interface{}) error {
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

	return nil
}

func (q *Query) Iterator(ctx context.Context, model Model) (*Iterator, error) {
	props, err := extractModelProps(model)
	if err != nil {
		return nil, err
	}

	cols := make([]string, len(props))
	for i, prop := range props {
		cols[i] = prop.Name
	}

	var conds []string
	var values []interface{}
	for _, cond := range q.conditions {
		conds = append(conds, cond.SQL())
		values = append(values, cond.Values()...)
	}

	sqlq := fmt.Sprintf(`SELECT %s FROM %s`, strings.Join(cols, ", "), model.TableName())
	if len(conds) > 0 {
		sqlq = fmt.Sprintf("%s WHERE %s", sqlq, strings.Join(conds, " AND "))
	}
	if len(q.orders) > 0 {
		sqlq = fmt.Sprintf("%s ORDER BY %s", sqlq, strings.Join(q.orders, ", "))
	}
	if q.limit > 0 {
		sqlq = fmt.Sprintf("%s LIMIT %d,%d", sqlq, q.offset, q.limit)
	}
	if isDebug() {
		log.Println("database [Iterator]:", sqlq)
	}

	rows, err := q.sess.QueryContext(ctx, sqlq, values...)
	if err != nil {
		return nil, err
	}

	return &Iterator{rows, props}, nil
}
