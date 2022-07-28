package crate

import (
	"database/sql/driver"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgtype"
)

func fieldName(fld reflect.StructField) string {
	if col, ok := fld.Tag.Lookup("db"); ok && col != "primary" {
		return strings.Split(col, ",")[0]
	}

	if col, ok := fld.Tag.Lookup("json"); ok {
		return strings.Split(col, ",")[0]
	}

	return fld.Name
}

func skipField(i any) bool {
	switch v := i.(type) {
	case IsZeroer:
		if v.IsZero() {
			return true
		}
	case pgtype.Text:
		if v.Status == pgtype.Undefined {
			return true
		}
	case pgtype.TextArray:
		if v.Status == pgtype.Undefined {
			return true
		}
	case pgtype.Timestamptz:
		if v.Status == pgtype.Undefined {
			return true
		}
	case driver.Valuer:
		if _, err := v.Value(); err != nil {
			return true
		}
	}

	return false
}

func writeInt[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](b *strings.Builder, v T) {
	b.Write(strconv.AppendInt([]byte{}, int64(v), 10))
}

func writeParam(b *strings.Builder, args *[]any, value any) {
	*args = append(*args, value)
	b.WriteByte('$')
	writeInt(b, len(*args))
}

func writeIdentifier(b *strings.Builder, identifiers ...string) {
	if len(identifiers) == 0 {
		return
	}

	first := true

	for _, id := range identifiers {
		if first {
			first = false
		} else {
			b.WriteString(", ")
		}

		b.WriteByte('"')
		b.WriteString(id)
		b.WriteByte('"')
	}
}
