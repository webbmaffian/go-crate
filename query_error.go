package crate

type QueryError struct {
	err   string
	query string
	args  []any
}

func (q QueryError) Error() string {
	return q.err
}

func (q QueryError) Query() string {
	return q.query
}

func (q QueryError) Args() []any {
	return q.args
}
