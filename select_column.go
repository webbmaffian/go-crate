package crate

import "strings"

type columns interface {
	writeColumns(b *strings.Builder)
	has(column string) bool
	count() int
	strings(cols *[]string)
}

type Columns []columns

func (c Columns) writeColumns(b *strings.Builder) {
	if len(c) == 0 {
		return
	}

	c[0].writeColumns(b)

	for _, v := range c[1:] {
		b.WriteString(", ")
		v.writeColumns(b)
	}
}

func (c Columns) has(column string) bool {
	for _, v := range c {
		if v.has(column) {
			return true
		}
	}

	return false
}

func (c Columns) count() (count int) {
	for _, v := range c {
		count += v.count()
	}

	return
}

func (c Columns) strings(cols *[]string) {
	for _, v := range c {
		v.strings(cols)
	}
}

type Column []string

func (c Column) writeColumns(b *strings.Builder) {
	if len(c) == 0 {
		return
	}

	writeIdentifier(b, c[0])

	for _, v := range c[1:] {
		b.WriteString(", ")
		writeIdentifier(b, v)
	}
}

func (c Column) has(column string) bool {
	for _, v := range c {
		if v == column {
			return true
		}
	}

	return false
}

func (c Column) count() int {
	return len(c)
}

func (c Column) strings(cols *[]string) {
	*cols = append(*cols, c...)
}

type AliasedColumn struct {
	Column string
	Alias  string
}

func (c AliasedColumn) writeColumns(b *strings.Builder) {
	writeIdentifier(b, c.Column)
	b.WriteString(" AS ")
	writeIdentifier(b, c.Alias)
}

func (c AliasedColumn) has(column string) bool {
	return c.Alias == column
}

func (c AliasedColumn) count() int {
	return 1
}

func (c AliasedColumn) strings(cols *[]string) {
	*cols = append(*cols, c.Alias)
}

type AggregatedColumn struct {
	Func         string
	ArgsCallback func(b *strings.Builder)
	Alias        string
}

func (c AggregatedColumn) writeColumns(b *strings.Builder) {
	if c.Func != "" {
		b.WriteString(c.Func)
	}

	if c.ArgsCallback != nil {
		b.WriteByte('(')
		c.ArgsCallback(b)
		b.WriteByte(')')
	}

	if c.Alias != "" {
		b.WriteString(" AS ")
		writeIdentifier(b, c.Alias)
	}
}

func (c AggregatedColumn) has(column string) bool {
	return c.Alias == column
}

func (c AggregatedColumn) count() int {
	return 1
}

func (c AggregatedColumn) strings(cols *[]string) {
	*cols = append(*cols, c.Alias)
}

func Count(alias string, distinctColumn ...string) AggregatedColumn {
	return AggregatedColumn{
		Func: "COUNT",
		ArgsCallback: func(b *strings.Builder) {
			if len(distinctColumn) == 0 {
				b.WriteString("*")
			} else {
				b.WriteString("DISTINCT ")
				writeIdentifier(b, distinctColumn[0])
			}
		},
		Alias: alias,
	}
}

func Unnest(column string, alias string) AggregatedColumn {
	return AggregatedColumn{
		Func: "UNNEST",
		ArgsCallback: func(b *strings.Builder) {
			writeIdentifier(b, column)
		},
		Alias: alias,
	}
}

func DateTrunc(per string, column string, alias string) AggregatedColumn {
	return AggregatedColumn{
		Func: "date_trunc",
		ArgsCallback: func(b *strings.Builder) {
			b.WriteByte('\'')
			b.WriteString(per)
			b.WriteString("', ")
			writeIdentifier(b, column)
		},
		Alias: alias,
	}
}

func Has(column string, alias string) AggregatedColumn {
	return AggregatedColumn{
		ArgsCallback: func(b *strings.Builder) {
			writeIdentifier(b, column)
			b.WriteString(" IS NOT NULL")
		},
		Alias: alias,
	}
}

func RawColumn(column string, alias ...string) AggregatedColumn {
	var a string

	if alias != nil {
		a = alias[0]
	}

	return AggregatedColumn{
		Func:  column,
		Alias: a,
	}
}

func Min(column string, alias string) AggregatedColumn {
	return Aggregate("MIN", column, alias)
}

func Max(column string, alias string) AggregatedColumn {
	return Aggregate("MAX", column, alias)
}

func Sum(column string, alias string) AggregatedColumn {
	return Aggregate("SUM", column, alias)
}

func ArrayAgg(column string, alias string) AggregatedColumn {
	return Aggregate("ARRAY_AGG", column, alias)
}

func Aggregate(aggFunc string, column string, alias string) AggregatedColumn {
	return AggregatedColumn{
		Func: aggFunc,
		ArgsCallback: func(b *strings.Builder) {
			writeIdentifier(b, column)
		},
		Alias: alias,
	}
}
