package crate

import (
	"context"
	"errors"
)

func (db *Crate) Iterate(ctx context.Context, q SelectQuery, iterator func(values []any) error) (err error) {
	if q.Select == nil {
		return errors.New("no columns to select")
	}

	err = q.run(ctx, db)

	if err != nil {
		return
	}

	defer q.Close()

	var values []any

	for q.Next() {
		if values, err = q.result.Values(); err != nil {
			return
		}

		if err = iterator(values); err != nil {
			return
		}
	}

	return
}

func (db *Crate) IterateRaw(ctx context.Context, q SelectQuery, iterator func(values [][]byte) error) (err error) {
	if q.Select == nil {
		return errors.New("no columns to select")
	}

	err = q.run(ctx, db)

	if err != nil {
		return
	}

	defer q.Close()

	for q.Next() {
		if err = iterator(q.result.RawValues()); err != nil {
			return
		}
	}

	return
}
