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

type Database struct {
	sess *sql.DB
}

func Open(ctx context.Context, credentials Credentials) (*Database, error) {
	if isDebug() {
		log.Println("database [Open]:", credentials)
	}

	sess, err := sql.Open("mysql", credentials.String())
	if err != nil {
		return nil, fmt.Errorf("database: cannot connect to mysql: %s", err)
	}

	sess.SetMaxOpenConns(3)
	sess.SetMaxIdleConns(0)

	if err := sess.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database: cannot ping mysql: %s", err)
	}

	return &Database{sess}, nil
}

func (db *Database) Get(ctx context.Context, model Model) error {
	props, err := extractModelProps(model)
	if err != nil {
		return err
	}

	cols := make([]string, len(props))
	pointers := make([]interface{}, len(props))
	conds := []string{}
	vconds := []interface{}{}
	for i, prop := range props {
		cols[i] = prop.Name
		pointers[i] = prop.Pointer

		if prop.PrimaryKey {
			conds = append(conds, fmt.Sprintf("%s = ?", prop.Name))
			vconds = append(vconds, prop.Value)
		}
	}

	q := fmt.Sprintf(`SELECT %s FROM %s WHERE %s`, strings.Join(cols, ", "), model.TableName(), strings.Join(conds, " AND "))
	if isDebug() {
		log.Println("database [Get]:", q)
	}

	if err := db.sess.QueryRowContext(ctx, q, vconds...).Scan(pointers...); err != nil {
		if err == sql.ErrNoRows {
			return ErrNoSuchEntity
		}

		return err
	}

	for _, prop := range props {
		prop.Value = reflect.ValueOf(prop.Pointer).Elem().Interface()
	}

	if h, ok := model.(ModelTrackingAfterGetHooker); ok {
		if err := h.ModelTrackingAfterGet(props); err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) Exec(ctx context.Context, query string, params ...interface{}) error {
	_, err := db.sess.ExecContext(ctx, query, params...)
	return err
}

func (db *Database) Close() {
	db.sess.Close()
}

func (db *Database) Put(ctx context.Context, model Model) error {
	props, err := extractModelProps(model)
	if err != nil {
		return err
	}

	var q string
	var values []interface{}
	var pks int
	if model.IsInserted() {
		var updates, cols, conds []string
		var vconds []interface{}
		for _, prop := range props {
			if prop.PrimaryKey {
				conds = append(conds, fmt.Sprintf("%s = ?", prop.Name))
				vconds = append(vconds, prop.Value)
				continue
			}

			if prop.OmitEmpty && isZero(prop.Value) {
				continue
			}

			cols = append(cols, prop.Name)
			updates = append(updates, fmt.Sprintf("%s = ?", prop.Name))
			values = append(values, prop.Value)
		}

		q = fmt.Sprintf(`UPDATE %s SET %s WHERE %s`, model.TableName(), strings.Join(updates, ", "), strings.Join(conds, " AND "))
		values = append(values, vconds...)
	} else {
		var placeholders, cols []string
		for _, prop := range props {
			if prop.PrimaryKey {
				if _, ok := prop.Value.(int64); ok {
					pks++
				}
			}

			if prop.OmitEmpty && isZero(prop.Value) {
				continue
			}

			cols = append(cols, prop.Name)
			placeholders = append(placeholders, "?")
			values = append(values, prop.Value)
		}

		q = fmt.Sprintf(`INSERT INTO %s(%s) VALUES(%s)`, model.TableName(), strings.Join(cols, ", "), strings.Join(placeholders, ", "))
	}
	if isDebug() {
		log.Println("database [Put]:", q)
	}

	result, err := db.sess.ExecContext(ctx, q, values...)
	if err != nil {
		return err
	}

	if pks == 1 && !model.IsInserted() {
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("database: cannot get last inserted id: %s", err)
		}

		v := reflect.ValueOf(model).Elem()
		v.FieldByName(getPrimaryKeyField(props)).Set(reflect.ValueOf(id))
	}

	if h, ok := model.(ModelTrackingAfterPutHooker); ok {
		if err := h.ModelTrackingAfterPut(props); err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) Filter(sql string, value interface{}) *Query {
	return db.FilterCond(&simpleCondition{sql, value})
}

func (db *Database) FilterCond(condition Condition) *Query {
	return newQuery(db, condition)
}

func (db *Database) GetAll(ctx context.Context, models interface{}) error {
	return newEmptyQuery(db).GetAll(ctx, models)
}

func (db *Database) Order(column string) *Query {
	return newEmptyQuery(db).Order(column)
}

func (db *Database) Count(ctx context.Context, model Model) (int64, error) {
	return newEmptyQuery(db).Count(ctx, model)
}

func (db *Database) Limit(limit int64) *Query {
	return newEmptyQuery(db).Limit(limit)
}

func (db *Database) Offset(offset int64) *Query {
	return newEmptyQuery(db).Offset(offset)
}

func (db *Database) Delete(ctx context.Context, model Model) error {
	props, err := extractModelProps(model)
	if err != nil {
		return err
	}

	var conds []string
	var values []interface{}
	for _, prop := range props {
		if !prop.PrimaryKey {
			continue
		}

		conds = append(conds, fmt.Sprintf("%s = ?", prop.Name))
		values = append(values, prop.Value)
	}

	q := fmt.Sprintf(`DELETE FROM %s WHERE %s`, model.TableName(), strings.Join(conds, " AND "))
	if isDebug() {
		log.Println("database [Delete]:", q)
	}

	if _, err := db.sess.ExecContext(ctx, q, values...); err != nil {
		return err
	}

	if h, ok := model.(ModelTrackingAfterDeleteHooker); ok {
		if err := h.ModelTrackingAfterDelete(props); err != nil {
			return err
		}
	}

	return nil
}
