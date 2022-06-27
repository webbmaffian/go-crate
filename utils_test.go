package crate

import (
	"testing"
)

func TestSanitizeFieldName(t *testing.T) {
	strs := []string{"bar", "foo.bar", "baz.foo.bar", "COUNT(*) AS bar"}

	for _, str := range strs {
		if result := sanitizeFieldName(str); result != "bar" {
			t.Error("Failed", str, "-", result, "is not 'bar'")
		}
	}
}
