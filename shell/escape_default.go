//go:build !windows

package shell

import "strings"

// EscapeChars replaces all double quotes and backslashes in the given string with escaped double quotes.
func EscapeChars(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, `"`, `\"`)
}
