package crate

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
)

type queryable interface {
	buildQuery(*[]any) string
}

type SelectQuery struct {
	Select  []string
	From    queryable
	Join    []Join
	Where   Condition
	GroupBy string
	Having  Condition
	OrderBy string
	Limit   int
	Offset  int

	result pgx.Rows
	args   *[]any
	error  error
}

type SelectOptions[T any] struct {
	BeforeMarshal func(*T) error
	AfterMarshal  func(*T) error
}

func (q *SelectQuery) Error() error {
	return q.error
}

func (q *SelectQuery) String() string {
	return q.buildQuery(&[]any{})
}

func (q *SelectQuery) buildQuery(args *[]any) string {
	q.args = args
	parts := make([]string, 0, 8)
	parts = append(parts, "SELECT "+strings.Join(q.Select, ", "))
	parts = append(parts, "FROM "+q.From.buildQuery(q.args))

	if q.Join != nil {
		for _, join := range q.Join {
			parts = append(parts, join.run(q.args))
		}
	}

	if q.Where != nil {
		parts = append(parts, "WHERE "+q.Where.run(q.args))
	}

	if q.GroupBy != "" {
		parts = append(parts, "GROUP BY "+q.GroupBy)
	}

	if q.Having != nil {
		parts = append(parts, "HAVING "+q.Having.run(q.args))
	}

	if q.OrderBy != "" {
		parts = append(parts, "ORDER BY "+q.OrderBy)
	}

	if q.Limit > 0 {
		parts = append(parts, "LIMIT "+strconv.Itoa(q.Limit))
	}

	if q.Offset > 0 {
		parts = append(parts, "OFFSET "+strconv.Itoa(q.Offset))
	}

	return strings.Join(parts, "\n")
}

func (q *SelectQuery) run(db *Crate) (err error) {
	q.result, err = db.pool.Query(context.Background(), q.String(), *q.args...)

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
