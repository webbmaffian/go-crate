package crate

import (
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