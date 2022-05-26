package crate

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var db Pgx

func SetConnection(conn *pgx.Conn) {
	registerJSONArrayType(conn)

	db = conn
}

func SetPool(pool *pgxpool.Pool) {
	pool.Config().AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		registerJSONArrayType(c)

		return nil
	}

	db = pool
}

// Register "JSON Array" (OID 199) type
func registerJSONArrayType(conn *pgx.Conn) {
	conn.ConnInfo().RegisterDataType(pgtype.DataType{
		Value: pgtype.NewArrayType("__json", pgtype.JSONOID, func() pgtype.ValueTranscoder { return &pgtype.JSON{} }),
		Name:  "__json",
		OID:   199,
	})
}
