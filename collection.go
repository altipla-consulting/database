package database

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Collection struct {
	sess          *sql.DB
	conditions    []Condition
	orders        []string
	offset, limit int64
	model         Model
	props         []*Property
}

func newCollection(db *Database, model Model) *Collection {
	props, err := extractModelProps(model)
	if err != nil {
		panic(err)
	}

	c := &Collection{
		sess:  db.sess,
		model: model,
		props: props,
	}

	return c
}

func (c *Collection) Clone() *Collection {
	return &Collection{
		sess:       c.sess,
		conditions: c.conditions,
		orders:     c.orders,
		offset:     c.offset,
		limit:      c.limit,
		model:      c.model,
		props:      c.props,
	}
}

func (c *Collection) Get(instance Model) error {
	modelProps := updatedProps(c.props, instance)
	b := &sqlBuilder{
		table:      c.model.TableName(),
		conditions: c.conditions,
		props:      modelProps,
	}

	for _, prop := range modelProps {
		if prop.PrimaryKey {
			b.conditions = append(b.conditions, &simpleCondition{prop.Name, prop.Value})
		}
	}

	statement, values := b.SelectSQL()
	if isDebug() {
		log.Println("database [Get]:", statement)
	}

	var pointers []interface{}
	for _, prop := range modelProps {
		pointers = append(pointers, prop.Pointer)
	}
	if err := c.sess.QueryRow(statement, values...).Scan(pointers...); err != nil {
		if err == sql.ErrNoRows {
			return ErrNoSuchEntity
		}

		return err
	}

	modelProps = updatedProps(c.props, instance)

	if err := instance.Tracking().AfterGet(modelProps); err != nil {
		return err
	}

	return nil
}

func (c *Collection) Put(instance Model) error {
	modelt := reflect.TypeOf(c.model)
	instancet := reflect.TypeOf(instance)
	if modelt != instancet {
		return fmt.Errorf("database: expected instance of %s and got a instance of %s", modelt, instancet)
	}

	b := &sqlBuilder{
		table: c.model.TableName(),
	}
	modelProps := updatedProps(c.props, instance)

	var q string
	var values []interface{}
	if instance.Tracking().IsInserted() {
		b.conditions = append(b.conditions, c.conditions...)
		b.conditions = append(b.conditions, &simpleCondition{"revision", instance.Tracking().StoredRevision()})

		for _, prop := range modelProps {
			if prop.PrimaryKey {
				b.conditions = append(b.conditions, &simpleCondition{prop.Name, prop.Value})
				continue
			}

			if prop.OmitEmpty && isZero(prop.Value) {
				continue
			}

			b.props = append(b.props, prop)
		}

		q, values = b.UpdateSQL()
	} else {
		for _, prop := range modelProps {
			if prop.OmitEmpty && isZero(prop.Value) {
				continue
			}

			b.props = append(b.props, prop)
		}

		q, values = b.InsertSQL()
	}
	if isDebug() {
		log.Println("database [Put]:", q)
	}

	result, err := c.sess.Exec(q, values...)
	if err != nil {
		return err
	}

	var pks int
	for _, prop := range modelProps {
		if prop.PrimaryKey {
			pks++
		}
	}
	if pks == 1 && !instance.Tracking().IsInserted() {
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("database: cannot get last inserted id: %s", err)
		}

		for _, prop := range modelProps {
			if prop.PrimaryKey {
				if _, ok := prop.Value.(int64); ok {
					reflect.ValueOf(prop.Pointer).Elem().Set(reflect.ValueOf(id))
				}
			}
		}
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("database: cannot get rows affected: %s", err)
	}
	if rows == 0 {
		return ErrConcurrentTransaction
	}

	if err := instance.Tracking().AfterPut(modelProps); err != nil {
		return err
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

func (c *Collection) Delete(instance Model) error {
	b := &sqlBuilder{
		table:      c.model.TableName(),
		conditions: c.conditions,
		limit:      1,
		offset:     c.offset,
	}
	modelProps := updatedProps(c.props, instance)

	for _, prop := range modelProps {
		if prop.PrimaryKey {
			b.conditions = append(b.conditions, &simpleCondition{prop.Name, prop.Value})
		}
	}

	statement, values := b.DeleteSQL()
	if isDebug() {
		log.Println("database [Delete]:", statement)
	}

	if _, err := c.sess.Exec(statement, values...); err != nil {
		return err
	}

	if err := instance.Tracking().AfterDelete(modelProps); err != nil {
		return err
	}

	return nil
}

func (c *Collection) Iterator() (*Iterator, error) {
	b := &sqlBuilder{
		table:      c.model.TableName(),
		conditions: c.conditions,
		props:      c.props,
		limit:      c.limit,
		offset:     c.offset,
		orders:     c.orders,
	}

	sql, values := b.SelectSQL()
	if isDebug() {
		log.Println("database [Iterator]:", sql)
	}

	rows, err := c.sess.Query(sql, values...)
	if err != nil {
		return nil, err
	}

	return &Iterator{rows, c.props}, nil
}

func (c *Collection) GetAll(models interface{}) error {
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

	modelt := reflect.TypeOf(c.model)
	if t.Elem() != modelt {
		return fmt.Errorf("database: expected a slice of %s and got a slice of %s", modelt, t.Elem())
	}

	dest := reflect.MakeSlice(t, 0, 0)

	it, err := c.Iterator()
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

func (c *Collection) First(instance Model) error {
	modelProps := updatedProps(c.props, instance)
	b := &sqlBuilder{
		table:      c.model.TableName(),
		conditions: c.conditions,
		props:      modelProps,
		limit:      c.limit,
		offset:     c.offset,
		orders:     c.orders,
	}

	statement, values := b.SelectSQL()
	if isDebug() {
		log.Println("database [First]:", statement)
	}

	var pointers []interface{}
	for _, prop := range modelProps {
		pointers = append(pointers, prop.Pointer)
	}
	if err := c.sess.QueryRow(statement, values...).Scan(pointers...); err != nil {
		if err == sql.ErrNoRows {
			return ErrNoSuchEntity
		}

		return err
	}

	modelProps = updatedProps(c.props, instance)

	if err := instance.Tracking().AfterGet(modelProps); err != nil {
		return err
	}

	return nil
}

func (c *Collection) Count() (int64, error) {
	b := &sqlBuilder{
		table:      c.model.TableName(),
		conditions: c.conditions,
	}

	sql, values := b.SelectSQLCols("COUNT(*)")
	if isDebug() {
		log.Println("database [Count]:", sql)
	}

	var n int64
	if err := c.sess.QueryRow(sql, values...).Scan(&n); err != nil {
		return 0, err
	}

	return n, nil
}

func (c *Collection) GetMulti(keys interface{}, models interface{}) error {
	v := reflect.ValueOf(models)
	t := reflect.TypeOf(models)
	keyst := reflect.TypeOf(keys)
	keysv := reflect.ValueOf(keys)

	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("database: pass a pointer to a slice of models to GetAll")
	}
	v = v.Elem()
	t = t.Elem()
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("database: pass a slice of models to GetAll")
	}

	if keyst.Kind() != reflect.Slice {
		return fmt.Errorf("database: pass a slice of keys to GetAll")
	}
	keyst = keyst.Elem()
	if keyst.Kind() != reflect.Int64 && keyst.Kind() != reflect.String {
		return fmt.Errorf("database: pass a slice of string/int64 keys to GetAll")
	}
	if keysv.Len() == 0 {
		return nil
	}

	var pk *Property
	for _, prop := range c.props {
		if prop.PrimaryKey {
			if pk != nil {
				return fmt.Errorf("database: cannot use GetMulti with multiple primary keys")
			}

			pk = prop
		}
	}

	c = c.Filter(fmt.Sprintf("%s IN", pk.Name), keys)

	fetch := reflect.New(t)
	fetch.Elem().Set(reflect.MakeSlice(t, 0, 0))
	if err := c.GetAll(fetch.Interface()); err != nil {
		return err
	}

	stringKeys := map[string]reflect.Value{}
	intKeys := map[int64]reflect.Value{}

	var merr MultiError
	for i := 0; i < fetch.Elem().Len(); i++ {
		model := fetch.Elem().Index(i)

		pk := model.Elem().FieldByName(pk.Field).Interface()
		switch v := pk.(type) {
		case string:
			stringKeys[v] = model
		case int64:
			intKeys[v] = model

		default:
			panic("should not reach here")
		}
	}

	results := reflect.MakeSlice(t, 0, 0)
	for i := 0; i < keysv.Len(); i++ {
		switch v := keysv.Index(i).Interface().(type) {
		case string:
			model, ok := stringKeys[v]
			if !ok {
				merr = append(merr, ErrNoSuchEntity)
				results = reflect.Append(results, reflect.Zero(t.Elem()))
				continue
			}

			merr = append(merr, nil)
			results = reflect.Append(results, model)

		case int64:
			model, ok := intKeys[v]
			if !ok {
				merr = append(merr, ErrNoSuchEntity)
				results = reflect.Append(results, reflect.Zero(t.Elem()))
				continue
			}

			merr = append(merr, nil)
			results = reflect.Append(results, model)

		default:
			panic("should not reach here")
		}
	}

	v.Set(results)

	if merr.HasError() {
		return merr
	}
	return nil
}

func (c *Collection) Truncate() error {
	b := &sqlBuilder{
		table: c.model.TableName(),
	}

	statement := b.TruncateSQL()
	if isDebug() {
		log.Println("database [Truncate]:", statement)
	}

	if _, err := c.sess.Exec(statement); err != nil {
		return err
	}

	statement = b.ResetAutoIncrementSQL()
	if isDebug() {
		log.Println("database [Truncate]:", statement)
	}

	if _, err := c.sess.Exec(statement); err != nil {
		return err
	}

	return nil
}
