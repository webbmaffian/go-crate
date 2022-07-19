package crate

import (
	"context"
	"errors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func NewCrate(config *pgxpool.Config) (c Crate, err error) {
	if config == nil {
		err = errors.New("Missing config")
		return
	}

	afterConnect := config.AfterConnect
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		registerDataTypes(conn)

		if afterConnect != nil {
			return afterConnect(ctx, conn)
		}

		return nil
	}

	c.pool, err = pgxpool.ConnectConfig(context.Background(), config)

	return
}

type Crate struct {
	pool *pgxpool.Pool
}

// Register "JSON Array" (OID 199) type
func registerDataTypes(conn *pgx.Conn) {
	conn.ConnInfo().RegisterDataType(pgtype.DataType{
		Value: pgtype.NewArrayType("__json", pgtype.JSONOID, func() pgtype.ValueTranscoder { return &pgtype.JSON{} }),
		Name:  "__json",
		OID:   199,
	})
}
