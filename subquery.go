package crate

func Subquery(alias string, query SelectQuery) SubquerySource {
	return SubquerySource{alias, query}
}

type SubquerySource struct {
	alias string
	query SelectQuery
}

func (t SubquerySource) buildQuery(args *[]any) string {
	return "(" + t.query.buildQuery(args) + ") AS " + t.alias
}
