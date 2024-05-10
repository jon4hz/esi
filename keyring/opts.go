package keyring

type Opt func(k *Keyring)

type KeyringType int

const (
	SessionKeyring KeyringType = iota + 1
	UserKeyring
)

func WithKeyringType(t KeyringType) Opt {
	return func(k *Keyring) {
		k.Type = t
	}
}
