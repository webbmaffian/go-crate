package crate

import (
	"context"
	"reflect"
	"strconv"
	"strings"
)

func (db *Crate) Update(table string, src any, condition Condition) (err error) {
	var fields []string
	var args []any

	switch v := src.(type) {

	case map[string]any:
		fields, args, err = updateFromMap(v)

	case *map[string]any:
		fields, args, err = updateFromMap(*v)

	case BeforeMutation:
		err = v.BeforeMutation(Updating)

		if err != nil {
			return
		}

		fields, args, err = updateFromStruct(src)

	case *BeforeMutation:
		err = (*v).BeforeMutation(Updating)

		if err != nil {
			return
		}

		fields, args, err = updateFromStruct(src)

	default:
		fields, args, err = updateFromStruct(src)
	}

	if err != nil {
		return
	}

	_, err = db.pool.Exec(context.Background(), "UPDATE "+table+" SET "+strings.Join(fields, ", ")+" WHERE "+condition.run(&args), args...)

	if err == nil {
		if s, ok := src.(AfterMutation); ok {
			s.AfterMutation(Updating)
		} else if s, ok := src.(*AfterMutation); ok {
			(*s).AfterMutation(Updating)
		}
	}

	return
}

func updateFromMap(src map[string]any) (fields []string, args []any, err error) {
	numFields := len(src)
	fields = make([]string, numFields)
	args = make([]any, numFields)

	i := 0

	for k, v := range src {
		fields[i] = k + " = $" + strconv.Itoa(i+1)
		args[i] = v
		i++
	}

	return
}

func updateFromStruct(src any) (fields []string, args []any, err error) {
	elem := reflect.ValueOf(src)

	if elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}

	typ := elem.Type()
	numFields := elem.NumField()
	fields = make([]string, numFields)
	args = make([]any, numFields)

	i := 0

	for idx := 0; idx < numFields; idx++ {
		f := elem.Field(idx)

		if !f.CanInterface() {
			continue
		}

		fld := typ.Field(idx)
		col := fieldName(fld)

		if fld.Tag.Get("db") == "primary" || fld.Tag.Get("db") == "-" {
			continue
		}

		v := f.Interface()

		if skipField(v) {
			continue
		}

		fields[i] = col + " = $" + strconv.Itoa(i+1)
		args[i] = v
		i++
	}

	return
}
