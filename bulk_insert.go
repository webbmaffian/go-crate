package crate

import (
	"context"
	"errors"
	"strings"
)

func (c *Crate) BulkInsert(ctx context.Context, table string, columns []string, rows [][]any, onConflict ...OnConflictUpdate) (err error) {
	numColumns := len(columns)

	if numColumns == 0 {
		return errors.New("No columns to insert")
	}

	numRows := len(rows)

	if numRows == 0 {
		return errors.New("No rows to insert")
	}

	var b strings.Builder
	args := make([]any, 0, numRows*numColumns)
	b.Grow(numColumns*16 + numColumns*4*numRows)

	b.WriteString("INSERT INTO ")
	writeIdentifier(&b, table)
	b.WriteString(" (")

	writeIdentifier(&b, columns[0])

	for _, col := range columns[1:] {
		b.WriteString(", ")
		writeIdentifier(&b, col)
	}

	b.WriteString(") VALUES ")

	firstRow := true

	for _, row := range rows {
		if len(row) != numColumns {
			return errors.New("Invalid number of columns")
		}

		if firstRow {
			firstRow = false
		} else {
			b.WriteString(",\n")
		}

		b.WriteByte('(')
		writeParam(&b, &args, row[0])

		for _, col := range row[1:] {
			writeParam(&b, &args, col)
		}

		b.WriteByte(')')
	}

	if len(onConflict) > 0 {
		if err = onConflict[0].run(&b, columns); err != nil {
			return
		}
	}

	_, err = c.pool.Exec(ctx, b.String(), args...)

	return
}
