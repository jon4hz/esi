//go:build windows

package shell

import (
	"os"
	"strings"
)

// EscapeChars replaces all double quotes and backslashes in the given string with escaped double quotes.
// If the SHELL variable isn't set, we assume that the user is running cyberark-ssh-utils from CMD or PowerShell.
// In this case, we don't need to escape quotes.
// If the user is running cyberark-ssh-utils from something like Git Bash, the SHELL variable will be set, and we need to escape quotes.
func EscapeChars(s string) string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return s
	}
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, `"`, `\"`)
}
