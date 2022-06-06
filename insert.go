package crate

import (
	"context"
	"reflect"
	"strconv"
	"strings"
)

func Insert(table string, src any) (err error) {
	elem := reflect.ValueOf(src)

	if elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}

	typ := elem.Type()
	numFields := elem.NumField()
	columns := make([]string, 0, numFields)
	placeholders := make([]string, 0, numFields)
	args := make([]any, 0, numFields)

	for i := 0; i < numFields; i++ {
		f := elem.Field(i)
		fld := typ.Field(i)
		col, ok := fld.Tag.Lookup("json")

		if !ok {
			col = fld.Name
		}

		columns = append(columns, col)
		placeholders = append(placeholders, "$"+strconv.Itoa(i+1))
		val := f.Interface()
		args = append(args, val)
	}

	_, err = db.Exec(context.Background(), "INSERT INTO "+table+" ("+strings.Join(columns, ", ")+") VALUES ("+strings.Join(placeholders, ", ")+")", args...)

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
