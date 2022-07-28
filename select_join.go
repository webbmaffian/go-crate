package crate

import "strings"

type join interface {
	runJoin(b *strings.Builder, args *[]any)
}

type Joins []join

func (j Joins) runJoin(b *strings.Builder, args *[]any) {
	for _, join := range j {
		join.runJoin(b, args)
	}
}

type InnerJoin struct {
	Table     queryable
	Condition Condition
}

func (j InnerJoin) runJoin(b *strings.Builder, args *[]any) {
	b.WriteString("INNER JOIN ")
	j.Table.buildQuery(b, args)
	b.WriteString(" ON ")
	j.Condition.run(b, args)
	b.WriteByte('\n')
}

type OuterJoin InnerJoin

func (j OuterJoin) runJoin(b *strings.Builder, args *[]any) {
	b.WriteString("OUTER JOIN ")
	j.Table.buildQuery(b, args)
	b.WriteString(" ON ")
	j.Condition.run(b, args)
	b.WriteByte('\n')
}

type LeftJoin InnerJoin

func (j LeftJoin) runJoin(b *strings.Builder, args *[]any) {
	b.WriteString("LEFT JOIN ")
	j.Table.buildQuery(b, args)
	b.WriteString(" ON ")
	j.Condition.run(b, args)
	b.WriteByte('\n')
}

type RightJoin InnerJoin

func (j RightJoin) runJoin(b *strings.Builder, args *[]any) {
	b.WriteString("RIGHT JOIN ")
	j.Table.buildQuery(b, args)
	b.WriteString(" ON ")
	j.Condition.run(b, args)
	b.WriteByte('\n')
}
