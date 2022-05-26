package crate

import (
	"context"
)

func Delete(table string, condition Condition) (err error) {
	args := make([]any, 0)
	_, err = db.Exec(context.Background(), "DELETE FROM "+table+" WHERE "+condition.run(&args), args...)

	return
}
