package crate

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

func (db *Crate) Select(ctx context.Context, dest any, q SelectQuery, options ...SelectOptions) (err error) {
	var opt SelectOptions

	if len(options) != 0 {
		opt = options[0]
	}

	switch d := dest.(type) {
	case io.Writer:
		return selectIntoWriter(ctx, d, &q, opt, db)

	case *map[string]any:
		return errors.New("Not supported yet")
	}

	destPtr := reflect.ValueOf(dest)

	if destPtr.Kind() != reflect.Pointer {
		return errors.New("Destination must be a pointer")
	}

	destVal := destPtr.Elem()

	switch destVal.Kind() {
	case reflect.Slice:
		err = selectIntoSlice(ctx, destPtr, &q, db)
	case reflect.Struct:
		err = selectOneIntoStruct(ctx, destPtr, &q, db)
	default:
		return errors.New("Invalid destination")
	}

	return
}

func selectOneIntoStruct(ctx context.Context, val reflect.Value, q *SelectQuery, db *Crate) (err error) {
	var selectedFields Column

	elem := val.Elem()
	typ := elem.Type()
	numFields := elem.NumField()
	destProps := make([]any, 0, numFields)
	q.Limit = 1

	if q.Select == nil {
		selectedFields = make(Column, 0, 10)
	}

	for i := 0; i < numFields; i++ {
		f := elem.Field(i)

		if !f.CanInterface() {
			continue
		}

		fld := typ.Field(i)
		col := fieldName(fld)

		if col == "-" {
			continue
		}

		if q.Select == nil || q.Select.has(col) {
			selectedFields = append(selectedFields, col)
			destProps = append(destProps, f.Addr().Interface())
		}
	}

	if q.Select == nil {
		q.Select = selectedFields
	}

	err = q.run(ctx, db)

	if err != nil {
		return
	}

	defer q.Close()

	var found bool

	for q.Next() {
		found = true
		err = q.Scan(destProps...)

		if err != nil {
			return
		}
	}

	if !found {
		err = errors.New("Row not found")
	}

	return
}

func selectIntoSlice(ctx context.Context, dest reflect.Value, q *SelectQuery, db *Crate) (err error) {
	var selectedFields Column

	destVal := dest.Elem()
	val := reflect.New(destVal.Type().Elem())
	elem := val.Elem()
	typ := elem.Type()
	numFields := elem.NumField()
	destProps := make([]any, 0, numFields)

	if q.Select == nil {
		selectedFields = make(Column, 0, 10)
	}

	for i := 0; i < numFields; i++ {
		f := elem.Field(i)

		if !f.CanInterface() {
			continue
		}

		fld := typ.Field(i)
		col := fieldName(fld)

		if col == "-" {
			continue
		}

		if q.Select == nil || q.Select.has(col) {
			selectedFields = append(selectedFields, col)
			destProps = append(destProps, f.Addr().Interface())
		}
	}

	err = q.run(ctx, db)

	if err != nil {
		return
	}

	defer q.Close()

	for q.Next() {
		err = q.Scan(destProps...)

		if err != nil {
			return
		}

		dest.Elem().Set(reflect.Append(destVal, elem))
	}

	return
}

func selectIntoWriter(ctx context.Context, w io.Writer, q *SelectQuery, opt SelectOptions, db *Crate) (err error) {
	if q.Select == nil {
		return errors.New("No columns to select")
	}

	err = q.run(ctx, db)

	if err != nil {
		return
	}

	defer q.Close()

	w.Write([]byte("["))

	var i int
	var b []byte
	var values []any
	numCols := q.Select.count()
	cols := make([]string, 0, numCols)
	m := make(map[string]any, numCols)

	q.Select.strings(&cols)

	for q.Next() {
		values, err = q.result.Values()

		if err != nil {
			return
		}

		for i, col := range cols {
			m[col] = values[i]
		}

		if opt.BeforeMarshal != nil {
			if err := opt.BeforeMarshal(&m); err != nil {
				continue
			}
		}

		if i != 0 {
			w.Write([]byte(","))
		}

		i++

		b, err = json.Marshal(m)

		if err != nil {
			return
		}

		_, err = w.Write(b)

		if err != nil {
			return
		}

		if opt.AfterMarshal != nil {
			opt.AfterMarshal(&m)
		}
	}

	w.Write([]byte("]"))

	return
}
