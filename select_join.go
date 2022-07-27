package crate

import "strings"

type Join interface {
	run(b *strings.Builder, args *[]any)
}

type InnerJoin struct {
	Table     queryable
	Condition Condition
}

func (j InnerJoin) run(b *strings.Builder, args *[]any) {
	b.WriteString("INNER JOIN ")
	j.Table.buildQuery(b, args)
	b.WriteString(" ON ")
	j.Condition.run(b, args)
}

type OuterJoin InnerJoin

func (j OuterJoin) run(b *strings.Builder, args *[]any) {
	b.WriteString("OUTER JOIN ")
	j.Table.buildQuery(b, args)
	b.WriteString(" ON ")
	j.Condition.run(b, args)
}

type LeftJoin InnerJoin

func (j LeftJoin) run(b *strings.Builder, args *[]any) {
	b.WriteString("LEFT JOIN ")
	j.Table.buildQuery(b, args)
	b.WriteString(" ON ")
	j.Condition.run(b, args)
}

type RightJoin InnerJoin

func (j RightJoin) run(b *strings.Builder, args *[]any) {
	b.WriteString("RIGHT JOIN ")
	j.Table.buildQuery(b, args)
	b.WriteString(" ON ")
	j.Condition.run(b, args)
}
