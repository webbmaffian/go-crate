package crate

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

func Insert(table string, src any, onConflict ...OnConflictUpdate) (err error) {
	elem := reflect.ValueOf(src)

	if elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}

	typ := elem.Type()
	numFields := elem.NumField()
	columns := make([]string, 0, numFields)
	placeholders := make([]string, 0, numFields)
	args := make([]any, 0, numFields)

	idx := 0

	for i := 0; i < numFields; i++ {
		f := elem.Field(i)
		fld := typ.Field(i)

		if !f.CanInterface() || fld.Tag.Get("db") == "-" {
			continue
		}

		col := fieldName(fld)

		idx++
		columns = append(columns, col)
		placeholders = append(placeholders, "$"+strconv.Itoa(idx))
		val := f.Interface()
		args = append(args, val)
	}

	q := "INSERT INTO " + table + " (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"

	if len(onConflict) > 0 {
		var str string

		str, err = onConflict[0].run(columns, placeholders)

		if err != nil {
			return
		}

		q += " " + str
	}

	_, err = db.Exec(context.Background(), q, args...)

	return
}

func InsertMultiple(table string, columns []string, rows [][]any, onConflict ...OnConflictUpdate) (err error) {
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

	_, err = db.Exec(context.Background(), q, values...)

	return
}

func BulkInsert(table string, columns []string, insert func() []any) (err error) {
	ctx := context.Background()
	poolConn, err := db.Acquire(ctx)

	if err != nil {
		return
	}

	defer poolConn.Release()

	conn := poolConn.Conn()
	numColumns := len(columns)
	placeholders := make([]string, 0, numColumns)

	for i := 0; i < numColumns; i++ {
		placeholders = append(placeholders, "$"+strconv.Itoa(i+1))
	}

	sd, err := conn.Prepare(ctx, "stmt_insert_"+table, "INSERT INTO "+table+" ("+strings.Join(columns, ", ")+") VALUES ("+strings.Join(placeholders, ", ")+")")

	if err != nil {
		return
	}

	defer conn.Deallocate(ctx, sd.Name)

	for {
		values := insert()

		if values == nil {
			break
		}

		_, err = conn.Exec(ctx, sd.Name, values...)

		if err != nil {
			break
		}
	}

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
