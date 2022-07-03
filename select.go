package crate

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
)

type queryable interface {
	buildQuery(*[]any) string
}

type SelectQuery struct {
	Select       []string
	From         string
	FromSubquery queryable
	Join         []Join
	Where        Condition
	GroupBy      string
	Having       Condition
	OrderBy      string
	Limit        int
	Offset       int

	result pgx.Rows
	args   *[]any
	error  error
}

type SelectOptions[T any] struct {
	BeforeMarshal func(*T) error
	AfterMarshal  func(*T) error
}

func (q *SelectQuery) Error() error {
	return q.error
}

func (q *SelectQuery) String() string {
	return q.buildQuery(&[]any{})
}

func (q *SelectQuery) buildQuery(args *[]any) string {
	q.args = args
	parts := make([]string, 0, 8)
	parts = append(parts, "SELECT "+strings.Join(q.Select, ", "))

	if q.FromSubquery != nil {
		parts = append(parts, "FROM ("+q.FromSubquery.buildQuery(q.args)+") AS subquery")
	} else {
		parts = append(parts, "FROM "+q.From)
	}

	if q.Join != nil {
		for _, join := range q.Join {
			parts = append(parts, join.run(q.args))
		}
	}

	if q.Where != nil {
		parts = append(parts, "WHERE "+q.Where.run(q.args))
	}

	if q.GroupBy != "" {
		parts = append(parts, "GROUP BY "+q.GroupBy)
	}

	if q.Having != nil {
		parts = append(parts, "HAVING "+q.Having.run(q.args))
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

	return strings.Join(parts, "\n")
}

func (q *SelectQuery) run() (err error) {
	if q.From == "" && q.FromSubquery == nil {
		return errors.New("Missing mandatory 'From' field")
	}

	q.result, err = db.Query(context.Background(), q.String(), *q.args...)

	return
}

func (q *SelectQuery) Next() bool {
	if q.result == nil {
		if q.error = q.run(); q.error != nil {
			return false
		}
	}

	n := q.result.Next()

	if !n {
		q.result = nil
	}

	return n
}

func (q *SelectQuery) Scan(dest ...any) error {
	if q.result == nil {
		return errors.New("Result is closed")
	}

	return q.result.Scan(dest...)
}

func (q *SelectQuery) Close() {
	if q.result != nil {
		q.result.Close()
		q.result = nil
	}
}

func SelectOne[T any](dest *T, q SelectQuery) (err error) {
	q.Limit = 1
	slice := make([]T, 0, 1)
	err = Select(&slice, q)

	if err != nil {
		return
	}

	if len(slice) == 0 {
		return errors.New("Row not found")
	}

	*dest = slice[0]

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

	err = q.run()

	if err != nil {
		return
	}

	defer q.Close()

	for q.Next() {
		err = q.Scan(destProps...)

		if err != nil {
			return
		}

		*dest = append(*dest, destStruct)
	}

	return
}

func SelectIntoJsonStream[T any](w io.Writer, destStruct T, q SelectQuery, options ...SelectOptions[T]) (err error) {
	var selectAll bool
	var opt SelectOptions[T]

	if len(options) > 0 {
		opt = options[0]
	}

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

		if !f.CanInterface() {
			continue
		}

		col := fieldName(fld)

		if selectAll {
			q.Select = append(q.Select, col)
		}

		if selectAll || containsSuffix(q.Select, col, "."+col, " "+col) {
			destProps = append(destProps, f.Addr().Interface())
		}
	}

	err = q.run()

	if err != nil {
		return
	}

	defer q.Close()

	w.Write([]byte("["))

	var i int
	var b []byte

	for q.Next() {
		err = q.Scan(destProps...)

		if err != nil {
			return
		}

		if opt.BeforeMarshal != nil {
			if err := opt.BeforeMarshal(&destStruct); err != nil {
				continue
			}
		}

		if i != 0 {
			w.Write([]byte(","))
		}

		i++

		b, err = json.Marshal(destStruct)

		if err != nil {
			return
		}

		_, err = w.Write(b)

		if err != nil {
			return
		}

		if opt.AfterMarshal != nil {
			opt.AfterMarshal(&destStruct)
		}
	}

	w.Write([]byte("]"))

	return
}

func SelectIntoArbitaryJsonStream(w io.Writer, q SelectQuery, options ...SelectOptions[map[string]any]) (err error) {
	var opt SelectOptions[map[string]any]

	if len(options) > 0 {
		opt = options[0]
	}

	err = q.run()

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
