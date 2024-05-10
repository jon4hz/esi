//go:build !linux
// +build !linux

package keyring

import "errors"

func loadKeyring(_ *Keyring) error { return errors.New("unsupported OS") }

func (k Keyring) Get(_ string) ([]byte, error) { return nil, nil }

func (k Keyring) GetAndRefresh(_ string, _ uint) ([]byte, error) { return nil, nil }

func (k Keyring) Store(_ string, _ []byte, _ uint) error { return nil }

func (k Keyring) Unlink(_ string) error { return nil }
