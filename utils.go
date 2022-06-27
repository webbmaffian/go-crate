package crate

import "strings"

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
