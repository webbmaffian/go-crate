package crate

import (
	"bytes"
	"strconv"
	"strings"
)

type Condition interface {
	run(args *[]any) string
}

type Raw string

func (c Raw) run(args *[]any) string {
	return string(c)
}

type Eq struct {
	Column string
	Value  any
}

func (c *Eq) run(args *[]any) string {
	*args = append(*args, c.Value)
	return c.Column + " = $" + strconv.Itoa(len(*args))
}

type NotEq struct {
	Column string
	Value  any
}

func (c *NotEq) run(args *[]any) string {
	*args = append(*args, c.Value)
	return c.Column + " != $" + strconv.Itoa(len(*args))
}

type Gt struct {
	Column string
	Value  any
}

func (c *Gt) run(args *[]any) string {
	*args = append(*args, c.Value)
	return c.Column + " > $" + strconv.Itoa(len(*args))
}

type Gte struct {
	Column string
	Value  any
}

func (c *Gte) run(args *[]any) string {
	*args = append(*args, c.Value)
	return c.Column + " >= $" + strconv.Itoa(len(*args))
}

type Lt struct {
	Column string
	Value  any
}

func (c *Lt) run(args *[]any) string {
	*args = append(*args, c.Value)
	return c.Column + " < $" + strconv.Itoa(len(*args))
}

type Lte struct {
	Column string
	Value  any
}

func (c *Lte) run(args *[]any) string {
	*args = append(*args, c.Value)
	return c.Column + " <= $" + strconv.Itoa(len(*args))
}

type And []Condition

func (c *And) run(args *[]any) string {
	conds := make([]string, 0, len(*c))

	for _, cond := range *c {
		conds = append(conds, cond.run(args))
	}

	return "(" + strings.Join(conds, " AND ") + ")"
}

type Or []Condition

func (c *Or) run(args *[]any) string {
	conds := make([]string, 0, len(*c))

	for _, cond := range *c {
		conds = append(conds, cond.run(args))
	}

	return "(" + strings.Join(conds, " OR ") + ")"
}

type In struct {
	Column string
	Value  any
}

func (c *In) run(args *[]any) (s string) {
	switch v := c.Value.(type) {
	case SelectQuery:
		s = c.Column + " IN (" + v.buildQuery(args) + ")"
	default:
		*args = append(*args, c.Value)
		s = c.Column + " = ANY $" + strconv.Itoa(len(*args))
	}

	return
}

func RawParams(str string, params ...any) (r *rawParams) {
	r = &rawParams{}
	r.String = []byte(str)
	r.Params = params

	return
}

type rawParams struct {
	String []byte
	Params []any
}

func (c *rawParams) run(args *[]any) (str string) {
	var prev int

	for _, param := range c.Params {
		cur := bytes.IndexByte(c.String[prev:], '?')

		if cur == -1 {
			break
		}

		*args = append(*args, param)
		str += string(c.String[prev:cur])
		str += "$" + strconv.Itoa(len(*args))

		prev = cur + 1
	}

	str += string(c.String[prev:])

	return
}
