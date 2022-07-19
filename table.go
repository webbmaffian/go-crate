package crate

func (c *Crate) Table(name string) TableSource {
	return TableSource{c, name}
}

type TableSource struct {
	db   *Crate
	name string
}

func (t TableSource) buildQuery(args *[]any) string {
	return t.name
}

func (t TableSource) Select(dest any, q SelectQuery, options ...SelectOptions[map[string]any]) error {
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
