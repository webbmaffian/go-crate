package crate

import (
	"database/sql/driver"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgtype"
)

var fieldRegex *regexp.Regexp

func init() {
	fieldRegex = regexp.MustCompile("([a-zA-Z0-9_]+)$")
}

func sanitizeFieldName(name any) string {
	return fieldRegex.FindString(name.(string))
}

func containsSuffix(slice []any, whole string, suffixes ...string) bool {
	for _, str := range slice {
		if str == whole {
			return true
		}

		for _, suffix := range suffixes {
			if v, ok := str.(string); ok {
				if strings.HasSuffix(v, suffix) {
					return true
				}
			}
		}
	}

	return false
}

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

func writeIdentifier(b *strings.Builder, identifiers ...any) {
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

		switch v := id.(type) {

		case string:
			b.WriteByte('"')
			b.WriteString(v)
			b.WriteByte('"')

		case RawString:
			b.WriteString(string(v))

		default:
			continue

		}
	}
}
