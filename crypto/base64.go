package crypto

import (
	"bytes"
	"encoding/base64"
)

func base64Encode(input []byte) []byte {
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(input)))
	base64.StdEncoding.Encode(encoded, input)
	return encoded
}

func base64Decode(input []byte) ([]byte, error) {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(input)))
	_, err := base64.StdEncoding.Decode(decoded, input)
	if err != nil {
		return nil, err
	}
	return bytes.Trim(decoded, "\x00"), nil
}
