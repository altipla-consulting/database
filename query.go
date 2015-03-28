package database

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/juju/errors"
)

// Query helps preparing and executing queries.
type Query struct {
	db *DB

	// Where conditions and replacement values
	conditions []string
	values     []interface{}

	limit int64
	order string
}

// Clone makes a copy of the query keeping all the internal state up to that moment.
func (q *Query) Clone() *Query {
	conditions := make([]string, len(q.conditions))
	for i := range conditions {
		conditions[i] = q.conditions[i]
	}
	values := make([]interface{}, len(q.values))
	for i := range values {
		values[i] = q.values[i]
	}

	return &Query{
		db:         q.db,
		conditions: conditions,
		values:     values,
		limit:      q.limit,
		order:      q.order,
	}
}

// Where filters by a new column adding some placeholders if needed.
func (q *Query) Where(column string, args ...interface{}) *Query {
	nargs := strings.Count(column, "?")
	if len(args) != nargs {
		panic(fmt.Sprintf("expected %d parameters in the query and received %d", nargs, len(args)))
	}
	if !hasOperator(column) {
		panic(fmt.Sprintf("column does not have an operator: %s", column))
	}

	if strings.HasSuffix(column, "IN (?)") && len(args) == 1 && reflect.TypeOf(args[0]).Kind() == reflect.Slice {
		argValue := reflect.ValueOf(args[0])
		placeholders := make([]string, argValue.Len())
		for i := range placeholders {
			placeholders[i] = "?"
		}
		newColumn := fmt.Sprintf("IN (%s)", strings.Join(placeholders, ", "))

		q.conditions = append(q.conditions, strings.Replace(column, "IN (?)", newColumn, -1))

		for i := 0; i < argValue.Len(); i++ {
			q.values = append(q.values, argValue.Index(i).Interface())
		}
	} else {
		q.conditions = append(q.conditions, column)
		q.values = append(q.values, args...)
	}

	return q
}

// GetAll returns all the results that matchs the query putting it in the output slice.
func (q *Query) GetAll(output interface{}) error {
	outputValue := reflect.ValueOf(output)
	outputType := reflect.TypeOf(output)

	// Some sanity checks about the output value
	if outputValue.Kind() != reflect.Ptr || outputValue.Elem().Kind() != reflect.Slice {
		return errors.New("output should be a pointer to a slice")
	}
	sliceElemType := outputType.Elem().Elem()
	if sliceElemType.Kind() != reflect.Ptr || sliceElemType.Elem().Kind() != reflect.Struct {
		return errors.New("output should be a pointer to a slice of struct pointers")
	}

	// Build the table name
	tableName := getTableName(sliceElemType.Elem())

	// Get the list of field names
	fields, columns, err := getSerializableFields(sliceElemType.Elem())
	if err != nil {
		return errors.Trace(err)
	}

	// Prepare the query string
	query := fmt.Sprintf("SELECT %s FROM `%s`", columns, tableName)
	if len(q.conditions) > 0 {
		query = fmt.Sprintf("%s WHERE %s", query, strings.Join(q.conditions, " AND "))
	}
	if q.order != "" {
		query = fmt.Sprintf("%s ORDER BY %s", query, q.order)
	}

	// Run the query and fetch the rows
	rows, err := q.db.Raw.Query(query, q.values...)
	if err != nil {
		return errors.Annotate(err, query)
	}
	defer rows.Close()

	scan := reflect.ValueOf(rows.Scan)
	for rows.Next() {
		// Create a new element
		elem := reflect.New(sliceElemType.Elem())

		// Prepare space to save the bytes of serialized fields
		serializedFields := [][]byte{}
		for _, field := range fields {
			if field.json || field.gob {
				serializedFields = append(serializedFields, []byte{})
			}
		}

		// Prepare the pointers to all its fields
		pointers := []reflect.Value{}
		serializedIdx := 0
		for _, field := range fields {
			if field.json || field.gob {
				pointers = append(pointers, reflect.ValueOf(&serializedFields[serializedIdx]))
				serializedIdx++
			} else {
				pointers = append(pointers, elem.Elem().FieldByName(field.name).Addr())
			}
		}

		// Scan the row into the struct
		if err := scan.Call(pointers)[0]; !err.IsNil() {
			return errors.Trace(err.Interface().(error))
		}

		// Read serialized fields
		serializedIdx = 0
		for _, field := range fields {
			switch {
			case field.json && len(serializedFields[serializedIdx]) > 0:
				decoder := json.NewDecoder(bytes.NewReader(serializedFields[serializedIdx]))
				dest := elem.Elem().FieldByName(field.name).Addr().Interface()
				if err := decoder.Decode(dest); err != nil {
					return errors.Trace(err)
				}

				serializedIdx++

			case field.gob:
				decoder := gob.NewDecoder(bytes.NewReader(serializedFields[serializedIdx]))
				dest := elem.Elem().FieldByName(field.name).Addr().Interface()
				if err := decoder.Decode(dest); err != nil {
					return errors.Trace(err)
				}

				serializedIdx++
			}
		}

		// Run hooks
		if err := runAfterFindHook(elem); err != nil {
			return errors.Trace(err)
		}

		// Append to the result
		outputValue.Elem().Set(reflect.Append(outputValue.Elem(), elem))
	}

	return nil
}

// Get returns the first result that matchs the query putting it in the output model.
func (q *Query) Get(output interface{}) error {
	outputValue := reflect.ValueOf(output)
	outputType := reflect.TypeOf(output)

	// Some sanity checks about the output value
	if outputValue.Kind() != reflect.Ptr || outputValue.Elem().Kind() != reflect.Struct {
		return errors.New("output should be a pointer to a struct")
	}

	// Limit the request to a single result
	q.Limit(1)

	// Buid an empty list for the results
	result := reflect.New(reflect.SliceOf(outputType))

	// Fetch the result
	getAll := reflect.ValueOf(q.GetAll)
	if err := getAll.Call([]reflect.Value{result})[0]; !err.IsNil() {
		return errors.Trace(err.Interface().(error))
	}

	resultElem := result.Elem()
	if resultElem.Len() == 0 {
		return ErrNoSuchEntity
	}

	// Output only the individual result, not the whole list
	outputValue.Elem().Set(resultElem.Index(0).Elem())

	return nil
}

// Limit returns only the specified number of results as a maximum
func (q *Query) Limit(limit int64) *Query {
	q.limit = limit

	return q
}

// Order sets the order of the rows in the result
func (q *Query) Order(order string) *Query {
	q.order = order

	return q
}

// Delete removes the models that match the query.
func (q *Query) Delete(model interface{}) error {
	modelValue := reflect.ValueOf(model)
	modelType := reflect.TypeOf(model)

	// Some sanity checks about the model
	if modelValue.Kind() != reflect.Ptr || modelValue.Elem().Kind() != reflect.Struct {
		return errors.New("model should be a pointer to a struct")
	}

	// Build the WHERE conditions of the query
	var conditions string
	if len(q.conditions) > 0 {
		conditions = fmt.Sprintf(" WHERE %s", strings.Join(q.conditions, " AND "))
	}

	// Build the table name
	tableName := getTableName(modelType.Elem())
	query := fmt.Sprintf("DELETE FROM `%s`%s", tableName, conditions)

	// Exec the query
	if q.db.Debug {
		log.Println("Delete:", query, "-->", q.values)
	}
	if _, err := q.db.Raw.Exec(query, q.values...); err != nil {
		return errors.Annotate(err, query)
	}

	return nil
}

func hasOperator(column string) bool {
	operators := []string{"=", "<>", "<", "<=", ">=", "IN"}
	for _, op := range operators {
		if strings.Contains(column, op) {
			return true
		}
	}

	return false
}
