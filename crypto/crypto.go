package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

type Msg struct {
	CipherText []byte
	Nonce      []byte
	Salt       []byte
}

const (
	kdfIterations = 250000
	kdfSaltLength = 8
	kdfKeySize    = 256
	adataSize     = 128
	cipherAlgo    = "aes"
	cipherMode    = "gcm"
)

func Encrypt(s []byte, passwd []byte) ([]byte, error) {
	kdfSalt, err := getRandomBytes(kdfSaltLength)
	if err != nil {
		return nil, fmt.Errorf("failed to get kdf salt: %w", err)
	}

	kdfKey := pbkdf2.Key(passwd, kdfSalt, kdfIterations, 32, sha256.New)

	nonce, err := getRandomBytes(16)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	block, err := aes.NewCipher(kdfKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, err
	}

	cipherText := aesgcm.Seal(nil, nonce, s, nil)

	msg := Msg{
		CipherText: cipherText,
		Salt:       kdfSalt,
		Nonce:      nonce,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return base64Encode(data), nil
}

func getRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

func Decrypt(cipherText []byte, passwd []byte) ([]byte, error) {
	raw, err := base64Decode(cipherText)
	if err != nil {
		return nil, err
	}

	var msg Msg
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, err
	}

	kdfKey := pbkdf2.Key(passwd, msg.Salt, kdfIterations, 32, sha256.New)

	block, err := aes.NewCipher(kdfKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, err
	}

	data, err := aesgcm.Open(nil, msg.Nonce, msg.CipherText, nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}
