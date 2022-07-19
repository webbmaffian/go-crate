package crate

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

func (db *Crate) Select(dest any, q SelectQuery, options ...SelectOptions[map[string]any]) (err error) {
	var opt SelectOptions[map[string]any]

	if len(options) != 0 {
		opt = options[0]
	}

	switch d := dest.(type) {
	case io.Writer:
		return selectIntoWriter(d, &q, opt, db)

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
		err = selectIntoSlice(destPtr, &q, db)
	case reflect.Struct:
		err = selectOneIntoStruct(destPtr, &q, db)
	default:
		return errors.New("Invalid destination")
	}

	if err != nil {
		err = QueryError{
			err:   err.Error(),
			query: q.String(),
			args:  *q.args,
		}
	}

	return
}

func selectOneIntoStruct(val reflect.Value, q *SelectQuery, db *Crate) (err error) {
	var selectAll bool
	elem := val.Elem()
	typ := elem.Type()
	numFields := elem.NumField()
	destProps := make([]any, 0, numFields)
	selectedFields := make([]string, len(q.Select))
	q.Limit = 1

	if len(q.Select) == 0 {
		selectAll = true
		q.Select = make([]string, 0, numFields)
	} else {
		copy(selectedFields, q.Select)
		q.Select = q.Select[:0]
	}

	for i := 0; i < numFields; i++ {
		f := elem.Field(i)

		if !f.CanInterface() {
			continue
		}

		fld := typ.Field(i)
		col := fieldName(fld)

		if selectAll || containsSuffix(selectedFields, col, "."+col, " "+col) {
			q.Select = append(q.Select, col)
			destProps = append(destProps, f.Addr().Interface())
		}
	}

	err = q.run(db)

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

func selectIntoSlice(dest reflect.Value, q *SelectQuery, db *Crate) (err error) {
	var selectAll bool

	destVal := dest.Elem()
	val := reflect.New(destVal.Type().Elem())
	elem := val.Elem()
	typ := elem.Type()
	numFields := elem.NumField()
	destProps := make([]any, 0, numFields)
	selectedFields := make([]string, len(q.Select))

	if len(q.Select) == 0 {
		selectAll = true
		q.Select = make([]string, 0, numFields)
	} else {
		copy(selectedFields, q.Select)
		q.Select = q.Select[:0]
	}

	for i := 0; i < numFields; i++ {
		f := elem.Field(i)

		if !f.CanInterface() {
			continue
		}

		fld := typ.Field(i)
		col := fieldName(fld)

		if selectAll || containsSuffix(selectedFields, col, "."+col, " "+col) {
			q.Select = append(q.Select, col)
			destProps = append(destProps, f.Addr().Interface())
		}
	}

	err = q.run(db)

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

func selectIntoWriter(w io.Writer, q *SelectQuery, opt SelectOptions[map[string]any], db *Crate) (err error) {
	err = q.run(db)

	if err != nil {
		return
	}

	defer q.Close()

	w.Write([]byte("["))

	var i int
	var b []byte
	var values []any
	cols := make([]string, 0, len(q.Select))
	m := map[string]any{}

	for _, col := range q.Select {
		cols = append(cols, sanitizeFieldName(col))
	}

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
