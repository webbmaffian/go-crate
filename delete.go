package crate

import (
	"context"
	"strings"
)

func (c *Crate) Delete(table string, condition Condition) (err error) {
	var b strings.Builder
	b.Grow(64)
	args := make([]any, 0, 2)

	b.WriteString("DELETE FROM ")
	writeIdentifier(&b, table)
	b.WriteByte('\n')
	b.WriteString("WHERE ")
	condition.run(&b, &args)

	_, err = c.pool.Exec(context.Background(), b.String(), args...)

	if err != nil {
		err = QueryError{
			err:   err.Error(),
			query: b.String(),
			args:  args,
		}
	}

	return
}
