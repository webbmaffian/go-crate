package crate

import (
	"context"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

func Update(table string, src any, condition Condition, columns ...string) (err error) {
	elem := reflect.ValueOf(src)

	if elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}

	allColumns := len(columns) == 0
	typ := elem.Type()
	numFields := elem.NumField()
	fields := make([]string, 0, numFields)
	args := make([]any, 0, numFields)

	idx := 0

	for i := 0; i < numFields; i++ {
		f := elem.Field(i)
		fld := typ.Field(i)
		col, ok := fld.Tag.Lookup("json")

		if !ok {
			col = fld.Name
		}

		if allColumns || slices.Contains(columns, col) {
			if fld.Tag.Get("db") == "primary" {
				continue
			}

			idx++
			fields = append(fields, col+" = $"+strconv.Itoa(idx))
			args = append(args, f.Interface())
		}
	}

	_, err = db.Exec(context.Background(), "UPDATE "+table+" SET "+strings.Join(fields, ", ")+" WHERE "+condition.run(&args), args...)

	return
}
