package main

import (
	"context"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

func UpdateRow(table string, src any, condition Map, columns ...string) (err error) {
	elem := reflect.ValueOf(src)

	if elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}

	allColumns := len(columns) == 0
	typ := elem.Type()
	numFields := elem.NumField()
	fields := make([]string, 0, numFields)
	numCond := len(condition)
	cond := make([]string, 0, numCond)
	args := make([]any, 0, numFields+numCond)

	idx := 0

	for i := 0; i < numFields; i++ {
		f := elem.Field(i)
		fld := typ.Field(i)
		col, ok := fld.Tag.Lookup("json")

		if !ok {
			col = fld.Name
		}

		if allColumns || slices.Contains(columns, col) {
			if _, exists := condition[col]; exists {
				continue
			}

			idx++
			fields = append(fields, col+" = $"+strconv.Itoa(idx))
			args = append(args, f.Interface())
		}
	}

	for k, v := range condition {
		idx++
		cond = append(cond, k+" = $"+strconv.Itoa(idx))
		args = append(args, v)
	}

	_, err = db.Exec(context.Background(), "UPDATE "+table+" SET "+strings.Join(fields, ", ")+" WHERE "+strings.Join(cond, ", "), args...)

	return
}
