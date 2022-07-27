package crate

import "strings"

type orderBy interface {
	orderBy(b *strings.Builder)
}

type OrderBy []orderBy

func (o OrderBy) orderBy(b *strings.Builder) {
	if len(o) == 0 {
		return
	}

	o[0].orderBy(b)

	for _, v := range o[1:] {
		b.WriteString(", ")
		v.orderBy(b)
	}
}

type Asc []string

func (o Asc) orderBy(b *strings.Builder) {
	if len(o) == 0 {
		return
	}

	b.WriteString(o[0])
	b.WriteString(" ASC")

	for _, v := range o[1:] {
		b.WriteString(", ")
		b.WriteString(v)
		b.WriteString(" ASC")
	}
}

type Desc []string

func (o Desc) orderBy(b *strings.Builder) {
	if len(o) == 0 {
		return
	}

	b.WriteString(o[0])
	b.WriteString(" DESC")

	for _, v := range o[1:] {
		b.WriteString(", ")
		b.WriteString(v)
		b.WriteString(" DESC")
	}
}
