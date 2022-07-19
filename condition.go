package crate

import (
	"bytes"
	"strconv"
	"strings"
)

type Condition interface {
	run(args *[]any) string
}

type Eq struct {
	Column string
	Value  any
}

func (c Eq) run(args *[]any) string {
	if c.Value == nil {
		return c.Column + " IS NULL"
	}

	*args = append(*args, c.Value)
	return c.Column + " = $" + strconv.Itoa(len(*args))
}

type NotEq struct {
	Column string
	Value  any
}

func (c NotEq) run(args *[]any) string {
	if c.Value == nil {
		return c.Column + " IS NOT NULL"
	}

	*args = append(*args, c.Value)
	return c.Column + " != $" + strconv.Itoa(len(*args))
}

type Gt struct {
	Column string
	Value  any
}

func (c Gt) run(args *[]any) string {
	*args = append(*args, c.Value)
	return c.Column + " > $" + strconv.Itoa(len(*args))
}

type Gte struct {
	Column string
	Value  any
}

func (c Gte) run(args *[]any) string {
	*args = append(*args, c.Value)
	return c.Column + " >= $" + strconv.Itoa(len(*args))
}

type Lt struct {
	Column string
	Value  any
}

func (c Lt) run(args *[]any) string {
	*args = append(*args, c.Value)
	return c.Column + " < $" + strconv.Itoa(len(*args))
}

type Lte struct {
	Column string
	Value  any
}

func (c Lte) run(args *[]any) string {
	*args = append(*args, c.Value)
	return c.Column + " <= $" + strconv.Itoa(len(*args))
}

type And []Condition

func (c And) run(args *[]any) string {
	conds := make([]string, 0, len(c))

	for _, cond := range c {
		conds = append(conds, cond.run(args))
	}

	return "(" + strings.Join(conds, " AND ") + ")"
}

type Or []Condition

func (c Or) run(args *[]any) string {
	conds := make([]string, 0, len(c))

	for _, cond := range c {
		conds = append(conds, cond.run(args))
	}

	return "(" + strings.Join(conds, " OR ") + ")"
}

type In struct {
	Column string
	Value  any
}

func (c In) run(args *[]any) (s string) {
	switch v := c.Value.(type) {
	case SelectQuery:
		s = c.Column + " IN (" + v.buildQuery(args) + ")"
	default:
		*args = append(*args, c.Value)
		s = c.Column + " = ANY $" + strconv.Itoa(len(*args))
	}

	return
}

func Raw(str string, params ...any) (r *raw) {
	r = &raw{}
	r.String = str
	r.Params = params

	return
}

type raw struct {
	String string
	Params []any
}

func (c raw) run(args *[]any) string {
	if len(c.Params) == 0 {
		return c.String
	}

	var str strings.Builder
	var prev int
	b := []byte(c.String)

	for _, param := range c.Params {
		cur := bytes.IndexByte(b[prev:], '?')

		if cur == -1 {
			break
		}

		*args = append(*args, param)
		str.Write(b[prev:cur])
		str.WriteString("$" + strconv.Itoa(len(*args)))

		prev = cur + 1
	}

	str.Write(b[prev:])

	return str.String()
}
