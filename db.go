package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/slices"
)

var db *pgx.Conn

func init() {
	var err error

	fmt.Println("Connecting to database...")

	connStr := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=disable", "crate", "", "localhost", "eriks_inventory")
	db, err = pgx.Connect(context.Background(), connStr)

	if err != nil {
		panic(err)
	}

	fmt.Println("Connected!")

	// Register "JSON Array" (OID 199) type
	db.ConnInfo().RegisterDataType(pgtype.DataType{
		Value: pgtype.NewArrayType("__json", pgtype.JSONOID, func() pgtype.ValueTranscoder { return &pgtype.JSON{} }),
		Name:  "__json",
		OID:   199,
	})
}

func SelectRow(table string, dest any, condition Map, columns ...string) (err error) {
	var selectAll bool

	elem := reflect.ValueOf(dest)

	if elem.Kind() != reflect.Pointer {
		return errors.New("SelectRow expects a pointer")
	}

	elem = elem.Elem()
	typ := elem.Type()
	numFields := elem.NumField()
	destProps := make([]any, 0, numFields)
	cond := make([]string, 0, len(condition))
	condArgs := make([]any, 0, len(condition))

	if len(columns) == 0 {
		selectAll = true
		columns = make([]string, 0, numFields)
	}

	for i := 0; i < numFields; i++ {
		f := elem.Field(i)
		fld := typ.Field(i)
		col, ok := fld.Tag.Lookup("json")

		if !ok {
			col = fld.Name
		}

		if selectAll {
			columns = append(columns, col)
		}

		if selectAll || slices.Contains(columns, col) {
			destProps = append(destProps, f.Addr().Interface())
		}
	}

	i := 0

	for k, v := range condition {
		i++
		cond = append(cond, k+" = $"+strconv.Itoa(i))
		condArgs = append(condArgs, v)
	}

	row := db.QueryRow(context.Background(), "SELECT "+strings.Join(columns, ", ")+" FROM "+table+" WHERE "+strings.Join(cond, ", "), condArgs...)

	return row.Scan(destProps...)
}

func InsertRow(table string, src any) (err error) {
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
