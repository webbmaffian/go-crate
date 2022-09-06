package crate

import (
	"context"
	"strings"
)

func (c *Crate) Table(name string) TableSource {
	return TableSource{c, name}
}

type TableSource struct {
	db   *Crate
	name string
}

func (t TableSource) buildQuery(b *strings.Builder, args *[]any) {
	writeIdentifier(b, t.name)
}

func (t TableSource) Select(dest any, q SelectQuery, options ...SelectOptions) error {
	q.From = t

	return t.db.Select(dest, q, options...)
}

func (t TableSource) Insert(src any, onConflict ...OnConflictUpdate) error {
	return t.db.Insert(t.name, src, onConflict...)
}

func (t TableSource) Update(src any, condition Condition) error {
	return t.db.Update(t.name, src, condition)
}

func (t TableSource) Delete(condition Condition) error {
	return t.db.Delete(t.name, condition)
}

func (t TableSource) Refresh() (err error) {
	var b strings.Builder
	b.Grow(100)

	b.WriteString("REFRESH TABLE ")
	writeIdentifier(&b, t.name)

	_, err = db.pool.Exec(context.Background(), b.String())

	return
}
