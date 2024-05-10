package shell

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeCharsWithShell(t *testing.T) {
	type testCase struct {
		input    string
		expected string
	}

	origShell := os.Getenv("SHELL")
	t.Cleanup(func() {
		os.Setenv("SHELL", origShell)
	})
	os.Setenv("SHELL", "/bin/bash")

	testCases := []testCase{
		{
			input:    `test`,
			expected: `test`,
		},
		{
			input:    `test"`,
			expected: `test\"`,
		},
		{
			input:    `test"test`,
			expected: `test\"test`,
		},
		{
			input:    `test"test""`,
			expected: `test\"test\"\"`,
		},
		{
			input:    `test"test"-'test'`,
			expected: `test\"test\"-'test'`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual := EscapeChars(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
