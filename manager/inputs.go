package manager

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/jon4hz/esi/crypto"
	"github.com/jon4hz/esi/forms"
)

const min15 = 900

func (m *Manager) Authenticate(forceNewPasswd, forceNewToken bool) error {
	token, err := m.gatherCredentials(forceNewPasswd, forceNewToken)
	if err != nil {
		return err
	}

	// now that we have a clear text token,
	// we try to connect to the secret server.
	if err := m.connectSecretServer(string(token)); err != nil {
		return fmt.Errorf("failed to connect to tss: %w", err)
	}

	return nil
}

func (m *Manager) gatherCredentials(forceNewPasswd, forceNewToken bool) ([]byte, error) {
	password, err := m.gatherPassword(forceNewPasswd)
	if err != nil {
		return nil, err
	}

	token, err := m.gatherToken(password, forceNewToken)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (m *Manager) gatherPassword(force bool) (password []byte, err error) {
	if !force {
		password, err = m.getPasswordFromKeyring()
		if err != nil {
			log.Warn("Failed to get password from keyring", "err", err)
		}
	}
	if err != nil || len(password) == 0 {
		password, err = m.getPasswordFromForm()
		if err != nil {
			return nil, err
		}
		if err := m.storePasswordInKeyring(password); err != nil {
			return nil, fmt.Errorf("failed to store password: %w", err)
		}
	}
	return
}

func (m *Manager) gatherToken(password []byte, force bool) (token []byte, err error) {
	// get the tss api token
	var tokenfromKeyring bool
	if !force {
		token, err = m.getTokenFromKeyring()
		if err != nil {
			log.Warn("Failed to get token from keyring", "err", err)
		}
	}
	if err != nil || len(token) == 0 {
		token, err = m.getTokenFromForm()
		if err != nil {
			return nil, err
		}
	} else {
		// if we can get the encrypted token from the keyring,
		// we'll use the password to decrypt it.
		var err error
		token, err = crypto.Decrypt(token, password)
		if err != nil {
			if err := m.sKeyring.Unlink(passwordID); err != nil {
				log.Warn("Failed to unlink faulty password", "err", err)
			}
			if strings.Contains(err.Error(), "message authentication failed") {
				log.Warn("Failed to decrypt token! Will retry...", "err", "wrong password")
				return m.gatherCredentials(true, false)
			}
			return nil, fmt.Errorf("failed to decrypt token: %w", err)
		}
		tokenfromKeyring = true
	}

	// only tmp.
	// we should store the token only if we are sure its working.
	if !tokenfromKeyring {
		encToken, err := crypto.Encrypt(token, password)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt token: %w", err)
		}
		if err := m.storeTokenInKeyring(encToken); err != nil {
			return nil, fmt.Errorf("failed to store token: %w", err)
		}
	}
	return
}

func (m *Manager) getTokenFromKeyring() ([]byte, error) {
	return m.uKeyring.Get(tokenID)
}

func (m *Manager) storeTokenInKeyring(token []byte) error {
	return m.uKeyring.Store(tokenID, token, m.cfg.SecretServer.TTL)
}

func (m *Manager) getTokenFromForm() ([]byte, error) {
	var token string
	if err := forms.TokenInputForm(&token).Run(); err != nil {
		return nil, fmt.Errorf("failed to get input: %w", err)
	}
	return []byte(token), nil
}

func (m *Manager) getPasswordFromKeyring() ([]byte, error) {
	return m.sKeyring.GetAndRefresh(passwordID, min15)
}

func (m *Manager) storePasswordInKeyring(password []byte) error {
	return m.sKeyring.Store(passwordID, password, min15)
}

func (m *Manager) getPasswordFromForm() ([]byte, error) {
	var password string
	if err := forms.PasswordInputForm(&password).Run(); err != nil {
		return nil, fmt.Errorf("failed to get input: %w", err)
	}
	return []byte(password), nil
}
