package crate

import (
	"testing"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DummyWriter struct{}

func (w DummyWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func BenchmarkSelect(B *testing.B) {
	var err error

	type User struct {
		Id        pgtype.Text `json:"id" db:"primary" validate:"required"`
		Status    pgtype.Text `json:"status" validate:"required,oneof=pending active inactive"`
		FirstName pgtype.Text `json:"first_name" validate:"required"`
		LastName  pgtype.Text `json:"last_name" validate:"required"`
	}

	B.Log("Connecting to database...")

	config, err := pgxpool.ParseConfig("postgresql://crate@localhost/test?sslmode=disable")

	if err != nil {
		B.Fatal(err)
	}

	db, err := NewCrate(config)

	if err != nil {
		B.Fatal(err)
	}

	defer db.pool.Close()

	B.ResetTimer()

	B.Run("Struct", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var dest User

			err = db.Select(&dest, SelectQuery{
				From: Table("users"),
			})

			if err != nil {
				b.Fatal(err)
			}
		}
	})

	B.Run("Slice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var dest []User

			err = db.Select(&dest, SelectQuery{
				From: Table("users"),
			})

			if err != nil {
				b.Fatal(err)
			}
		}
	})

	B.Run("Writer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dest := DummyWriter{}

			err = db.Select(dest, SelectQuery{
				Select: []string{"id", "status", "first_name", "last_name"},
				From:   Table("users"),
			})

			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
