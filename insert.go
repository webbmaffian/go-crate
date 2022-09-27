package crate

import (
	"context"
	"reflect"
	"strings"
)

func (db *Crate) Insert(ctx context.Context, table string, src any, onConflict ...OnConflictUpdate) (err error) {
	var b strings.Builder
	b.Grow(200)

	var values strings.Builder
	b.Grow(100)

	keys := make([]string, 0, 10)
	args := make([]any, 0, 10)

	b.WriteString("INSERT INTO ")
	writeIdentifier(&b, table)
	b.WriteByte(' ')

	switch v := src.(type) {

	case map[string]any:
		err = insertFromMap(&b, &values, &keys, &args, v)

	case *map[string]any:
		err = insertFromMap(&b, &values, &keys, &args, *v)

	case BeforeMutation:
		err = v.BeforeMutation(Inserting)

		if err != nil {
			return
		}

		err = insertFromStruct(&b, &values, &keys, &args, src)

	case *BeforeMutation:
		err = (*v).BeforeMutation(Inserting)

		if err != nil {
			return
		}

		err = insertFromStruct(&b, &values, &keys, &args, src)

	default:
		err = insertFromStruct(&b, &values, &keys, &args, src)

	}

	if err != nil {
		return
	}

	b.WriteByte('\n')
	b.WriteString("VALUES ")
	b.WriteString(values.String())

	if len(onConflict) > 0 {
		if err = onConflict[0].run(&b, keys); err != nil {
			return
		}
	}

	_, err = db.pool.Exec(ctx, b.String(), args...)

	if err == nil {
		if s, ok := src.(AfterMutation); ok {
			s.AfterMutation(Inserting)
		} else if s, ok := src.(*AfterMutation); ok {
			(*s).AfterMutation(Inserting)
		}
	} else {
		err = QueryError{
			err:   err.Error(),
			query: b.String(),
			args:  args,
		}
	}

	return
}

func insertFromMap(b *strings.Builder, values *strings.Builder, keys *[]string, args *[]any, src map[string]any) (err error) {
	first := true

	b.WriteByte('(')
	values.WriteByte('(')

	for k, v := range src {
		if first {
			first = false
		} else {
			b.WriteString(", ")
			values.WriteString(", ")
		}

		writeIdentifier(b, k)
		writeParam(values, args, v)
		*keys = append(*keys, k)
	}

	b.WriteByte(')')
	values.WriteByte(')')

	return
}

func insertFromStruct(b *strings.Builder, values *strings.Builder, keys *[]string, args *[]any, src any) (err error) {
	first := true

	b.WriteByte('(')
	values.WriteByte('(')

	elem := reflect.ValueOf(src)

	if elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}

	typ := elem.Type()
	numFields := elem.NumField()

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

		if first {
			first = false
		} else {
			b.WriteString(", ")
			values.WriteString(", ")
		}

		writeIdentifier(b, col)
		writeParam(values, args, v)
		*keys = append(*keys, col)
	}

	b.WriteByte(')')
	values.WriteByte(')')

	return
}
