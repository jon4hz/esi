// go:build linux
// build +linux

package keyring

import (
	"errors"
	"fmt"

	"github.com/jon4hz/keyctl"
)

func loadKeyring(k *Keyring) error {
	var err error
	switch k.Type {
	case SessionKeyring:
		k.keyring, err = keyctl.SessionKeyring()
		return err
	case UserKeyring:
		k.keyring, err = keyctl.UserKeyring()
		return err
	default:
		return errors.New("unsupported keyring type")
	}
}

func (k Keyring) Get(id string) ([]byte, error) {
	key, err := k.keyring.Search(id)
	if err != nil {
		return nil, err
	}
	data, err := key.Get()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (k Keyring) GetAndRefresh(id string, ttl uint) ([]byte, error) {
	key, err := k.keyring.Search(id)
	if err != nil {
		return nil, fmt.Errorf("failed to search for key: %w", err)
	}
	data, err := key.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}
	if err := key.ExpireAfter(ttl); err != nil {
		return nil, fmt.Errorf("failed to refresh ttl: %w", err)
	}
	return data, nil
}

func (k Keyring) Store(id string, value []byte, ttl uint) error {
	key, err := k.keyring.Add(id, value)
	if err != nil {
		return err
	}
	key.ExpireAfter(ttl) //nolint:errcheck

	return nil
}

func (k Keyring) Unlink(id string) error {
	key, err := k.keyring.Search(id)
	if err != nil {
		return err
	}
	return key.Unlink()
}
