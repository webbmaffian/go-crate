package crate

import (
	"context"
)

func (c *Crate) Delete(table string, condition Condition) (err error) {
	args := make([]any, 0)
	q := "DELETE FROM " + table + " WHERE " + condition.run(&args)
	_, err = c.pool.Exec(context.Background(), q, args...)

	if err != nil {
		err = QueryError{
			err:   err.Error(),
			query: q,
			args:  args,
		}
	}

	return
}
