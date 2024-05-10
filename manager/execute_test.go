package manager

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildExecCmd(t *testing.T) {
	type testCase struct {
		input    []string
		expected string
	}

	origShell := os.Getenv("SHELL")
	t.Cleanup(func() {
		os.Setenv("SHELL", origShell)
	})
	os.Setenv("SHELL", "/bin/bash")

	testCases := []testCase{
		{
			input:    []string{"test"},
			expected: `test`,
		},
		{
			input:    []string{"ls", "-l"},
			expected: `ls -l`,
		},
		{
			input:    []string{"echo", `"this is a test"`},
			expected: `echo \"this is a test\"`,
		},
		{
			input:    []string{"echo", `"this is a test with \"quotes\""`},
			expected: `echo \"this is a test with \\\"quotes\\\"\"`,
		},
		{
			input:    []string{"echo", `\"`, "something", `\"`},
			expected: `echo \\\" something \\\"`,
		},
		{
			input:    []string{"echo", `\'`, "something", `\'`},
			expected: `echo \\' something \\'`,
		},
	}

	m := Manager{}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			actual := m.buildExecCmd(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
