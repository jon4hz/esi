package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase64Encoding(t *testing.T) {
	for _, tc := range []string{
		"simple",
		"test_1",
		"%^+;[",
		"prv-asdf",
		"aasdf41q245äö",
		"123442546262",
		"sadf*g45\"+",
	} {
		encoded := base64Encode([]byte(tc))
		decoded, err := base64Decode(encoded)
		assert.NoError(t, err)
		assert.Equal(t, []byte(tc), decoded)
	}
}

func FuzzBase64Encoding(f *testing.F) {
	for _, tc := range []string{
		"simple",
		"test_1",
		"%^+;[",
		"prv-asdf",
		"aasdf41q245äö",
		"123442546262",
		"sadf*g45\"+",
	} {
		f.Add([]byte(tc))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		if bytes.HasSuffix(raw, []byte("\x00")) {
			t.Skip("Skip special case where test has zero byte suffix.")
		}
		encoded := base64Encode(raw)
		decoded, err := base64Decode(encoded)
		assert.NoError(t, err)
		assert.Equal(t, raw, decoded)
	})
}
