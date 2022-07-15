package crate

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

func (db *Crate) Query(sql string, args ...any) (pgx.Rows, error) {
	return db.pool.Query(context.Background(), sql, args...)
}

func (db *Crate) QueryRow(sql string, args ...any) pgx.Row {
	return db.pool.QueryRow(context.Background(), sql, args...)
}

func (db *Crate) Exec(sql string, args ...any) (pgconn.CommandTag, error) {
	return db.pool.Exec(context.Background(), sql, args...)
}

func (db *Crate) CtxQuery(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return db.pool.Query(ctx, sql, args...)
}

func (db *Crate) CtxQueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return db.pool.QueryRow(ctx, sql, args...)
}

func (db *Crate) CtxExec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, sql, args...)
}
