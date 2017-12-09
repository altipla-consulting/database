package database

import (
	"database/sql"
)

type Iterator struct {
	rows  *sql.Rows
	props []*Property
}

func (it *Iterator) Close() {
	it.rows.Close()
}

func (it *Iterator) Next(model Model) error {
	modelProps := updatedProps(it.props, model)

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

	ptrs := make([]interface{}, len(modelProps))
	for i, prop := range modelProps {
		ptrs[i] = prop.Pointer
	}
	if err := it.rows.Scan(ptrs...); err != nil {
		return err
	}

	modelProps = updatedProps(it.props, model)

	if err := model.Tracking().ModelTrackingAfterGet(modelProps); err != nil {
		return err
	}

	return nil
}
