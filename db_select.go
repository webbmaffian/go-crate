package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
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
		parts = append(parts, "WHERE "+q.Where.run(args))
	}

	if q.GroupBy != "" {
		parts = append(parts, "GROUP BY "+q.GroupBy)
	}

	if q.Having != nil {
		parts = append(parts, "HAVING "+q.Having.run(args))
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

// Fill destination with the results
func (q *SelectQuery) Fill(dest any) (err error) {
	var destProps []any
	var destStructPtr reflect.Value
	destPtr := reflect.ValueOf(&dest)

	if destPtr.Kind() != reflect.Pointer {
		return errors.New("Destination must be a pointer")
	}

	elem := destPtr.Elem()

	switch t := elem.Kind(); t {
	case reflect.Slice:
		sliceType := elem.Type()
		sliceElem := sliceType.Elem()
		numFields := sliceElem.NumField()
		destProps = make([]any, 0, numFields)
		q.Select = make([]string, 0, numFields)

		destStructPtr = reflect.New(sliceElem)
		destStruct := destStructPtr.Elem()

		for i := 0; i < numFields; i++ {
			f := destStruct.Field(i)
			fld := sliceElem.Field(i)
			col, ok := fld.Tag.Lookup("json")

			if !ok {
				col = fld.Name
			}

			q.Select = append(q.Select, col)
			destProps = append(destProps, f.Addr().Interface())

			// columns = append(columns, col)
			// placeholders = append(placeholders, "$"+strconv.Itoa(i+1))
			// val := f.Interface()
			// args = append(args, val)
		}
	case reflect.Pointer:
	default:
		return errors.New("Destination must be either slice or pointer to a struct. Provided: " + t.String())
	}

	fmt.Println("Running query...")

	err = q.run()

	if err != nil {
		return
	}

	defer q.result.Close()

	for q.result.Next() {
		q.result.Scan(destProps...)

		// fmt.Println(elem.Kind())

		destPtr.Set(reflect.Append(elem, destStructPtr))
	}

	return
}

type Condition interface {
	run(args []any) string
}

type Raw string

func (c *Raw) run(args []any) string {
	return string(*c)
}

type Eq struct {
	Column string
	Value  any
}

func (c *Eq) run(args []any) string {
	args = append(args, c.Value)
	return c.Column + " = $" + strconv.Itoa(len(args))
}

type NotEq struct {
	Column string
	Value  any
}

func (c *NotEq) run(args []any) string {
	args = append(args, c.Value)
	return c.Column + " != $" + strconv.Itoa(len(args))
}

type Gt struct {
	Column string
	Value  any
}

func (c *Gt) run(args []any) string {
	args = append(args, c.Value)
	return c.Column + " > $" + strconv.Itoa(len(args))
}

type Gte struct {
	Column string
	Value  any
}

func (c *Gte) run(args []any) string {
	args = append(args, c.Value)
	return c.Column + " >= $" + strconv.Itoa(len(args))
}

type Lt struct {
	Column string
	Value  any
}

func (c *Lt) run(args []any) string {
	args = append(args, c.Value)
	return c.Column + " > $" + strconv.Itoa(len(args))
}

type Lte struct {
	Column string
	Value  any
}

func (c *Lte) run(args []any) string {
	args = append(args, c.Value)
	return c.Column + " >= $" + strconv.Itoa(len(args))
}

type And []Condition

func (c *And) run(args []any) string {
	conds := make([]string, 0, len(*c))

	for _, cond := range *c {
		conds = append(conds, cond.run(args))
	}

	return "(" + strings.Join(conds, " AND ") + ")"
}

type Or []Condition

func (c *Or) run(args []any) string {
	conds := make([]string, 0, len(*c))

	for _, cond := range *c {
		conds = append(conds, cond.run(args))
	}

	return "(" + strings.Join(conds, " OR ") + ")"
}
