// Package stringutils is a package for string utility methods
package stringutils

import "strings"

func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
