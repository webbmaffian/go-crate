package crate

import (
	"context"
)

func (c *Crate) Delete(table string, condition Condition) (err error) {
	args := make([]any, 0)
	_, err = c.pool.Exec(context.Background(), "DELETE FROM "+table+" WHERE "+condition.run(&args), args...)

	return
}
