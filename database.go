package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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

type Model interface {
	TableName() string
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

	return nil
}

func (db *Database) Exec(ctx context.Context, query string, params ...interface{}) error {
	_, err := db.sess.ExecContext(ctx, query, params...)
	return err
}

func (db *Database) Close() {
	db.sess.Close()
}
