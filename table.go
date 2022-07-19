package crate

func Table(name string) TableSource {
	return TableSource{name}
}

type TableSource struct {
	name string
}

func (t TableSource) buildQuery(args *[]any) string {
	return t.name
}
