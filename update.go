package crate

import (
	"context"
	"database/sql/driver"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgtype"
	"golang.org/x/exp/slices"
)

func Update(table string, src any, condition Condition, columns ...string) (err error) {
	if s, ok := src.(BeforeMutation); ok {
		err = s.BeforeMutation(Updating)
	} else if s, ok := src.(*BeforeMutation); ok {
		err = (*s).BeforeMutation(Updating)
	}

	if err != nil {
		return
	}

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

		if !f.CanInterface() {
			continue
		}

		fld := typ.Field(i)
		col := fieldName(fld)

		if allColumns || slices.Contains(columns, col) {
			if fld.Tag.Get("db") == "primary" || fld.Tag.Get("db") == "-" {
				continue
			}

			i := f.Interface()

			switch v := i.(type) {
			case pgtype.Text:
				if v.Status == pgtype.Undefined {
					continue
				}
			case pgtype.Timestamptz:
				if v.Status == pgtype.Undefined {
					continue
				}
			case driver.Valuer:
				if _, err = v.Value(); err != nil {
					continue
				}
			}

			idx++
			fields = append(fields, col+" = $"+strconv.Itoa(idx))
			args = append(args, i)
		}
	}

	_, err = db.Exec(context.Background(), "UPDATE "+table+" SET "+strings.Join(fields, ", ")+" WHERE "+condition.run(&args), args...)

	if s, ok := src.(AfterMutation); ok {
		s.AfterMutation(Updating)
	} else if s, ok := src.(*AfterMutation); ok {
		(*s).AfterMutation(Updating)
	}

	return
}
