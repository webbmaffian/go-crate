package crate

import (
	"context"
	"reflect"
	"strconv"
	"strings"
)

func (db *Crate) Insert(table string, src any, onConflict ...OnConflictUpdate) (err error) {
	var columns []string
	var placeholders []string
	args := make([]any, 0, 10)

	switch v := src.(type) {

	case map[string]any:
		columns, placeholders, err = insertFromMap(v, &args)

	case *map[string]any:
		columns, placeholders, err = insertFromMap(*v, &args)

	case BeforeMutation:
		err = v.BeforeMutation(Updating)

		if err != nil {
			return
		}

		columns, placeholders, err = insertFromStruct(src, &args)

	case *BeforeMutation:
		err = (*v).BeforeMutation(Updating)

		if err != nil {
			return
		}

		columns, placeholders, err = insertFromStruct(src, &args)

	default:
		columns, placeholders, err = insertFromStruct(src, &args)

	}

	if err != nil {
		return
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

	_, err = db.pool.Exec(context.Background(), q, args...)

	if err == nil {
		if s, ok := src.(AfterMutation); ok {
			s.AfterMutation(Inserting)
		} else if s, ok := src.(*AfterMutation); ok {
			(*s).AfterMutation(Inserting)
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

func insertFromMap(src map[string]any, args *[]any) (columns []string, placeholders []string, err error) {
	numFields := len(src)
	columns = make([]string, numFields)
	placeholders = make([]string, numFields)

	i := 0

	for k, v := range src {
		columns[i] = k
		placeholders[i] = "$" + strconv.Itoa(i+1)
		*args = append(*args, v)
		i++
	}

	return
}

func insertFromStruct(src any, args *[]any) (columns []string, placeholders []string, err error) {
	elem := reflect.ValueOf(src)

	if elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}

	typ := elem.Type()
	numFields := elem.NumField()
	columns = make([]string, numFields)
	placeholders = make([]string, numFields)
	i := 0

	for idx := 0; idx < numFields; idx++ {
		fld := typ.Field(idx)
		val := elem.Field(idx)

		if !val.CanInterface() || fld.Tag.Get("db") == "-" {
			continue
		}

		v := val.Interface()

		if skipField(v) {
			continue
		}

		col := fieldName(fld)

		columns[i] = col
		placeholders[i] = "$" + strconv.Itoa(i+1)
		*args = append(*args, v)
		i++
	}

	return
}
