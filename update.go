package crate

import (
	"context"
	"reflect"
	"strconv"
	"strings"
)

func (db *Crate) Update(table string, src any, condition Condition) (err error) {
	var fields []string
	args := make([]any, 0, 10)

	switch v := src.(type) {

	case map[string]any:
		fields, err = updateFromMap(v, &args)

	case *map[string]any:
		fields, err = updateFromMap(*v, &args)

	case BeforeMutation:
		err = v.BeforeMutation(Updating)

		if err != nil {
			return
		}

		fields, err = updateFromStruct(src, &args)

	case *BeforeMutation:
		err = (*v).BeforeMutation(Updating)

		if err != nil {
			return
		}

		fields, err = updateFromStruct(src, &args)

	default:
		fields, err = updateFromStruct(src, &args)
	}

	if err != nil {
		return
	}

	q := "UPDATE " + table + " SET " + strings.Join(fields, ", ") + " WHERE " + condition.run(&args)
	_, err = db.pool.Exec(context.Background(), q, args...)

	if err == nil {
		if s, ok := src.(AfterMutation); ok {
			s.AfterMutation(Updating)
		} else if s, ok := src.(*AfterMutation); ok {
			(*s).AfterMutation(Updating)
		}
	} else {
		err = QueryError{
			err:   err.Error(),
			query: q,
			args:  args,
		}
	}

	return
}

func updateFromMap(src map[string]any, args *[]any) (fields []string, err error) {
	numFields := len(src)
	fields = make([]string, numFields)

	i := 0

	for k, v := range src {
		fields[i] = k + " = $" + strconv.Itoa(i+1)
		*args = append(*args, v)
		i++
	}

	return
}

func updateFromStruct(src any, args *[]any) (fields []string, err error) {
	elem := reflect.ValueOf(src)

	if elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}

	typ := elem.Type()
	numFields := elem.NumField()
	fields = make([]string, 0, numFields)

	i := 0

	for idx := 0; idx < numFields; idx++ {
		f := elem.Field(idx)

		if !f.CanInterface() {
			continue
		}

		fld := typ.Field(idx)
		col := fieldName(fld)

		if fld.Tag.Get("db") == "primary" || col == "-" {
			continue
		}

		v := f.Interface()

		if skipField(v) {
			continue
		}

		fields = append(fields, col+" = $"+strconv.Itoa(i+1))
		*args = append(*args, v)
		i++
	}

	return
}
