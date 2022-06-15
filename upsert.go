package crate

import (
	"context"
	"strconv"
	"strings"
)

type BulkedUpsert struct {
	query string
}

func (b *BulkedUpsert) Values(values ...any) (err error) {
	_, err = db.Exec(context.Background(), b.query, values...)

	return
}

func UpsertBulk(table string, uniqueColumn string, otherColumns ...string) BulkedUpsert {
	numColumns := len(otherColumns)
	insertPlaceholders := make([]string, 0, numColumns+1)
	updatePlaceholders := make([]string, 0, numColumns)

	insertPlaceholders = append(insertPlaceholders, "$1")

	for i := 0; i < numColumns; i++ {
		insertPlaceholders = append(insertPlaceholders, "$"+strconv.Itoa(i+2))
		updatePlaceholders = append(updatePlaceholders, otherColumns[i]+" = $"+strconv.Itoa(i+2))
	}

	return BulkedUpsert{
		query: "INSERT INTO " + table + "(" + uniqueColumn + ", " + strings.Join(otherColumns, ", ") + ") VALUES(" + strings.Join(insertPlaceholders, ", ") + ") ON CONFLICT (" + uniqueColumn + ") DO UPDATE SET " + strings.Join(updatePlaceholders, ", ") + ";",
	}
}
