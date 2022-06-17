package crate

import (
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
		v.args = args
		s = c.Column + " IN (" + v.buildQuery() + ")"
	default:
		*args = append(*args, c.Value)
		s = c.Column + " = ANY $" + strconv.Itoa(len(*args))
	}

	return
}
