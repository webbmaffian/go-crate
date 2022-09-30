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

func (t TableSource) Select(ctx context.Context, dest any, q SelectQuery, options ...SelectOptions) error {
	q.From = t

	return t.db.Select(ctx, dest, q, options...)
}

func (t TableSource) Iterate(ctx context.Context, q SelectQuery, iterator func(values []any) error) error {
	q.From = t

	return t.db.Iterate(ctx, q, iterator)
}

func (t TableSource) IterateRaw(ctx context.Context, q SelectQuery, iterator func(values [][]byte) error) error {
	q.From = t

	return t.db.IterateRaw(ctx, q, iterator)
}

func (t TableSource) Insert(ctx context.Context, src any, onConflict ...OnConflictUpdate) error {
	return t.db.Insert(ctx, t.name, src, onConflict...)
}

func (t TableSource) Update(ctx context.Context, src any, condition Condition) error {
	return t.db.Update(ctx, t.name, src, condition)
}

func (t TableSource) Delete(ctx context.Context, condition Condition) error {
	return t.db.Delete(ctx, t.name, condition)
}

func (t TableSource) Refresh(ctx context.Context) (err error) {
	var b strings.Builder
	b.Grow(100)

	b.WriteString("REFRESH TABLE ")
	writeIdentifier(&b, t.name)

	_, err = t.db.pool.Exec(ctx, b.String())

	return
}
