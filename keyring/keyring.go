package keyring

import (
	"errors"

	"github.com/jon4hz/keyctl"
)

type Keyring struct {
	Type    KeyringType
	keyring keyctl.Keyring
}

func New(opts ...Opt) (*Keyring, error) {
	k := &Keyring{
		Type: SessionKeyring,
	}
	for _, o := range opts {
		o(k)
	}

	if err := loadKeyring(k); err != nil {
		return nil, err
	}
	return k, nil
}

var ErrKeyringUninitialized = errors.New("keyring not initialized")
