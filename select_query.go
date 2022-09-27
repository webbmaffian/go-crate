package crate

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v4"
)

type queryable interface {
	buildQuery(b *strings.Builder, args *[]any)
}

type SelectQuery struct {
	Select  columns
	From    queryable
	Join    join
	Where   Condition
	GroupBy columns
	Having  Condition
	OrderBy orderBy
	Limit   int
	Offset  int

	result pgx.Rows
	error  error
}

type SelectOptions struct {
	BeforeMarshal func(data any) error
	AfterMarshal  func(data any) error
}

func (q *SelectQuery) Error() error {
	return q.error
}

func (q *SelectQuery) String() string {
	var b strings.Builder
	q.buildQuery(&b, &[]any{})
	return b.String()
}

func (q *SelectQuery) buildQuery(b *strings.Builder, args *[]any) {
	b.Grow(300)

	if q.Select != nil {
		b.WriteString("SELECT ")
		q.Select.writeColumns(b)
		b.WriteByte('\n')
	}

	if q.From != nil {
		b.WriteString("FROM ")
		q.From.buildQuery(b, args)
		b.WriteByte('\n')
	}

	if q.Join != nil {
		q.Join.runJoin(b, args)
	}

	if q.Where != nil {
		b.WriteString("WHERE ")
		q.Where.run(b, args)
		b.WriteByte('\n')
	}

	if q.GroupBy != nil {
		b.WriteString("GROUP BY ")
		q.GroupBy.writeColumns(b)
		b.WriteByte('\n')
	}

	if q.Having != nil {
		b.WriteString("HAVING ")
		q.Having.run(b, args)
		b.WriteByte('\n')
	}

	if q.OrderBy != nil {
		b.WriteString("ORDER BY ")
		q.OrderBy.orderBy(b)
		b.WriteByte('\n')
	}

	if q.Limit > 0 {
		b.WriteString("LIMIT ")
		writeInt(b, q.Limit)
		b.WriteByte('\n')
	}

	if q.Offset > 0 {
		b.WriteString("OFFSET ")
		writeInt(b, q.Offset)
		b.WriteByte('\n')
	}
}

func (q *SelectQuery) run(ctx context.Context, db *Crate) (err error) {
	var b strings.Builder
	args := make([]any, 0, 5)
	q.buildQuery(&b, &args)

	q.result, err = db.pool.Query(ctx, q.String(), args...)

	if err != nil {
		err = QueryError{
			err:   err.Error(),
			query: q.String(),
			args:  args,
		}
	}

	return
}

func (q *SelectQuery) Next() bool {
	if q.result == nil {
		return false
	}

	n := q.result.Next()

	if !n {
		q.result = nil
	}

	return n
}

func (q *SelectQuery) Scan(dest ...any) error {
	if q.result == nil {
		return errors.New("Result is closed")
	}

	return q.result.Scan(dest...)
}

func (q *SelectQuery) Close() {
	if q.result != nil {
		q.result.Close()
		q.result = nil
	}
}
