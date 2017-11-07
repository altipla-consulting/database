package database

import (
	"database/sql"
	"errors"
	"log"
)

type Database struct {
	sess *sql.DB
}

func NewDatabase(dsn string) (*Database, error) {
	sess, err := sql.Open("mysql", dsn)
	if err != nil {
	  return nil, err
	}

	sess.SetMaxOpenConns(3)
	sess.SetMaxIdleConns(0)

	return &Database{dsn}, nil
}

type Model interface {
	TableName() string
}

func (db *Database) Get(ctx context.Context, model Model) error {
	props, err := extractModelProps(model)
	if err != nil {
	  return errors.Trace(err)
	}

	cols := make([]string, len(props))
	for i, prop := range props {
		cols[i] = prop.Name
	}

	q := fmt.Sprintf(`SELECT %s FROM %s WHERE %s`, strings.Join(cols, ", "), model.TableName(), conds)
	if isDebug() {
		log.Println("database [Get]:", q)
	}

	db.sess.QueryRowContext(ctx, q)
	return nil
}
