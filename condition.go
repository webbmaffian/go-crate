package crate

import (
	"bytes"
	"strings"
)

type Condition interface {
	run(b *strings.Builder, args *[]any)
}

type Eq struct {
	Column string
	Value  any
}

func (c Eq) run(b *strings.Builder, args *[]any) {
	writeIdentifier(b, c.Column)

	if c.Value == nil {
		b.WriteString(" IS NULL")
	} else {
		b.WriteString(" = ")
		writeParam(b, args, c.Value)
	}
}

type NotEq struct {
	Column string
	Value  any
}

func (c NotEq) run(b *strings.Builder, args *[]any) {
	writeIdentifier(b, c.Column)

	if c.Value == nil {
		b.WriteString(" IS NOT NULL")
	} else {
		b.WriteString(" != ")
		writeParam(b, args, c.Value)
	}
}

type Gt struct {
	Column string
	Value  any
}

func (c Gt) run(b *strings.Builder, args *[]any) {
	writeIdentifier(b, c.Column)
	b.WriteString(" > ")
	writeParam(b, args, c.Value)
}

type Gte struct {
	Column string
	Value  any
}

func (c Gte) run(b *strings.Builder, args *[]any) {
	writeIdentifier(b, c.Column)
	b.WriteString(" >= ")
	writeParam(b, args, c.Value)
}

type Lt struct {
	Column string
	Value  any
}

func (c Lt) run(b *strings.Builder, args *[]any) {
	writeIdentifier(b, c.Column)
	b.WriteString(" < ")
	writeParam(b, args, c.Value)
}

type Lte struct {
	Column string
	Value  any
}

func (c Lte) run(b *strings.Builder, args *[]any) {
	writeIdentifier(b, c.Column)
	b.WriteString(" <= ")
	writeParam(b, args, c.Value)
}

type And []Condition

func (c And) run(b *strings.Builder, args *[]any) {
	if len(c) == 0 {
		return
	}

	b.WriteByte('(')
	c[0].run(b, args)

	for _, cond := range c[1:] {
		b.WriteString(" AND ")
		cond.run(b, args)
	}

	b.WriteByte(')')
}

type Or []Condition

func (c Or) run(b *strings.Builder, args *[]any) {
	if len(c) == 0 {
		return
	}

	b.WriteByte('(')
	c[0].run(b, args)

	for _, cond := range c[1:] {
		b.WriteString(" OR ")
		cond.run(b, args)
	}

	b.WriteByte(')')
}

type In struct {
	Column string
	Value  any
}

func (c In) run(b *strings.Builder, args *[]any) {
	writeIdentifier(b, c.Column)

	switch v := c.Value.(type) {
	case SelectQuery:
		b.WriteString(" IN (")
		v.buildQuery(b, args)
		b.WriteByte(')')
	case SubquerySource:
		b.WriteString(" IN (")
		v.query.buildQuery(b, args)
		b.WriteByte(')')
	default:
		b.WriteString(" = ANY ")
		writeParam(b, args, c.Value)
	}

	return
}

type NotIn struct {
	Column string
	Value  any
}

func (c NotIn) run(b *strings.Builder, args *[]any) {
	writeIdentifier(b, c.Column)

	switch v := c.Value.(type) {
	case SelectQuery:
		b.WriteString(" NOT IN (")
		v.buildQuery(b, args)
		b.WriteByte(')')
	case SubquerySource:
		b.WriteString(" NOT IN (")
		v.query.buildQuery(b, args)
		b.WriteByte(')')
	default:
		b.WriteString(" != ANY ")
		writeParam(b, args, c.Value)
	}

	return
}

type Contains struct {
	Column string
	Value  any
}

func (c Contains) run(b *strings.Builder, args *[]any) {
	writeParam(b, args, c.Value)
	b.WriteString(" = ANY ")
	writeIdentifier(b, c.Column)
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

func (c raw) run(b *strings.Builder, args *[]any) {
	if len(c.Params) == 0 {
		b.WriteString(c.String)
		return
	}

	var prev int
	str := []byte(c.String)

	for _, param := range c.Params {
		cur := bytes.IndexByte(str[prev:], '?')

		if cur == -1 {
			break
		}

		b.Write(str[prev:cur])
		writeParam(b, args, param)

		prev = cur + 1
	}

	b.Write(str[prev:])
}
