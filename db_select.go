package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/slices"
)

type SelectQuery struct {
	Select  []string
	From    string
	Where   Condition
	GroupBy string
	Having  Condition
	OrderBy string
	Limit   int
	Offset  int

	result pgx.Rows
}

func (q *SelectQuery) run() (err error) {
	// TODO: Reflect dest instead
	if len(q.Select) == 0 {
		q.Select = []string{"*"}
	}

	if q.From == "" {
		return errors.New("Missing mandatory 'From' field")
	}

	var args []any
	parts := make([]string, 0, 6)
	parts = append(parts, "SELECT "+strings.Join(q.Select, ", "))
	parts = append(parts, "FROM "+q.From)

	if q.Where != nil {
		parts = append(parts, "WHERE "+q.Where.run(&args))
	}

	if q.GroupBy != "" {
		parts = append(parts, "GROUP BY "+q.GroupBy)
	}

	if q.Having != nil {
		parts = append(parts, "HAVING "+q.Having.run(&args))
	}

	if q.OrderBy != "" {
		parts = append(parts, "ORDER BY "+q.OrderBy)
	}

	if q.Limit > 0 {
		parts = append(parts, "LIMIT "+strconv.Itoa(q.Limit))
	}

	if q.Offset > 0 {
		parts = append(parts, "OFFSET "+strconv.Itoa(q.Offset))
	}

	query := strings.Join(parts, "\n") + ";"
	fmt.Println(query)
	q.result, err = db.Query(context.Background(), query, args...)

	return
}

func (q *SelectQuery) Test(dest any) (err error) {
	v := reflect.ValueOf(dest)

	v.Set(reflect.Append(v, reflect.ValueOf(123)))

	return
}

func SelectOne[T any](dest *T, q SelectQuery) (err error) {
	q.Limit = 1
	slice := make([]T, 0, 1)
	err = Select(&slice, q)

	if err == nil {
		*dest = slice[0]
	}

	return
}

func Select[T any](dest *[]T, q SelectQuery) (err error) {
	var destStruct T
	var selectAll bool

	val := reflect.ValueOf(&destStruct)
	elem := val.Elem()
	typ := elem.Type()
	numFields := elem.NumField()
	destProps := make([]any, 0, numFields)

	if len(q.Select) == 0 {
		selectAll = true
		q.Select = make([]string, 0, numFields)
	}

	for i := 0; i < numFields; i++ {
		f := elem.Field(i)
		fld := typ.Field(i)
		col, ok := fld.Tag.Lookup("json")

		if !ok {
			col = fld.Name
		}

		if selectAll {
			q.Select = append(q.Select, col)
		}

		if selectAll || slices.Contains(q.Select, col) {
			destProps = append(destProps, f.Addr().Interface())
		}
	}

	err = q.run()

	if err != nil {
		return
	}

	defer q.result.Close()

	for q.result.Next() {
		err = q.result.Scan(destProps...)

		if err != nil {
			return
		}

		*dest = append(*dest, destStruct)
	}

	return
}
