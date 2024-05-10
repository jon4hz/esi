package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncrypt(t *testing.T) {
	_, err := Encrypt([]byte("MyClearText09876"), []byte("abc123SuperPassword"))
	assert.NoError(t, err)
}

func TestEncryptUnique(t *testing.T) {
	results := make([][]byte, 20)
	for i := 0; i < 20; i++ {
		s, err := Encrypt([]byte("MyClearText09876"), []byte("123456"))
		assert.NoError(t, err)
		assert.NotContains(t, results, s)
		results[i] = s
	}
}

func TestDecrypt(t *testing.T) {
	e, err := Encrypt([]byte("MyClearText09876"), []byte("abc123SuperPassword"))
	assert.NoError(t, err)

	d, err := Decrypt(e, []byte("abc123SuperPassword"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("MyClearText09876"), d)

}

func FuzzEncryption(f *testing.F) {
	var password = []byte("HelloThere$<")
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
		encrypted, err := Encrypt(raw, password)
		assert.NoError(t, err)
		decrypted, err := Decrypt(encrypted, password)
		assert.NoError(t, err)
		assert.Equal(t, raw, decrypted)
	})
}

func FuzzEncryptionPassword(f *testing.F) {
	var text = []byte("HelloThere$<")
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
		encrypted, err := Encrypt(text, raw)
		assert.NoError(t, err)
		decrypted, err := Decrypt(encrypted, raw)
		assert.NoError(t, err)
		assert.Equal(t, text, decrypted)
	})
}
