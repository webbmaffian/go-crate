package crate

import (
	"reflect"
	"regexp"
	"strings"
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
