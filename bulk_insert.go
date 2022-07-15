package crate

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

func (c *Crate) BulkInsert(table string, columns []string, rows [][]any, onConflict ...OnConflictUpdate) (err error) {
	numRows := len(rows)

	if numRows == 0 {
		return errors.New("No rows to insert")
	}

	numColumns := len(rows[0])
	placeholders := make([]string, numColumns)
	values := make([]any, 0, numRows*numColumns)
	idx := 0
	q := "INSERT INTO " + table + " (" + strings.Join(columns, ", ") + ") VALUES "
	first := true

	for _, row := range rows {
		if len(row) != numColumns {
			return errors.New("Invalid number of columns")
		}

		for i := range placeholders {
			idx++
			placeholders[i] = "$" + strconv.Itoa(idx)
		}

		if !first {
			q += ", "
		}

		q += "(" + strings.Join(placeholders, ", ") + ")"
		values = append(values, row...)

		first = false
	}

	if len(onConflict) > 0 {
		var str string

		str, err = onConflict[0].run(columns, placeholders)

		if err != nil {
			return
		}

		q += " " + str
	}

	_, err = c.pool.Exec(context.Background(), q, values...)

	return
}

type OnConflictUpdate []string

func (conflictingColumns OnConflictUpdate) run(columns []string, placeholders []string) (str string, err error) {
	numCols := len(columns)

	if numCols != len(placeholders) {
		err = errors.New("Length of columns and placeholders mismatch")
	}

	values := make([]string, 0, numCols)

	for _, column := range columns {
		if slices.Contains(conflictingColumns, column) {
			continue
		}

		values = append(values, column+" = excluded."+column)
	}

	str = "ON CONFLICT (" + strings.Join(conflictingColumns, ", ") + ") DO UPDATE SET " + strings.Join(values, ", ")

	return
}
