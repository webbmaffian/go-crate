package crate

import (
	"database/sql/driver"
	"reflect"
	"regexp"
	"strings"

	"github.com/jackc/pgtype"
)

var fieldRegex *regexp.Regexp

func init() {
	fieldRegex = regexp.MustCompile("([a-zA-Z0-9_]+)$")
}

func sanitizeFieldName(name string) string {
	return fieldRegex.FindString(name)
}

func containsSuffix(slice []string, whole string, suffixes ...string) bool {
	for _, str := range slice {
		if str == whole {
			return true
		}

		for _, suffix := range suffixes {
			if strings.HasSuffix(str, suffix) {
				return true
			}
		}
	}

	return false
}

func fieldName(fld reflect.StructField) string {
	if col, ok := fld.Tag.Lookup("db"); ok && col != "-" && col != "primary" {
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
