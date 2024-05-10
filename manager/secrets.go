package manager

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/jon4hz/esi/config"
	"github.com/jon4hz/tss-sdk-go/v2/server"
)

func (m *Manager) connectSecretServer(token string) error {
	srvCfg := server.Configuration{
		ServerURL: m.cfg.SecretServer.URL,
		Credentials: server.UserCredential{
			Token: token,
		},
	}
	srv, err := server.New(srvCfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	m.server = srv
	return nil
}

func (m *Manager) requiredSecrets(inj *config.Injector) []*config.Secret {
	requiredSecrets := make([]*config.Secret, 0)
	for _, c := range inj.Configs {
		if c.EnvSecret != "" {
			secret := m.cfg.SecretByID(c.EnvSecret)
			if secret != nil {
				requiredSecrets = append(requiredSecrets, secret)
			}
		}
		if c.StdoutSecret != "" {
			secret := m.cfg.SecretByID(c.StdoutSecret)
			if secret != nil {
				requiredSecrets = append(requiredSecrets, secret)
			}
		}
		if len(c.TmpFileSecrets) != 0 {
			for _, s := range c.TmpFileSecrets {
				secret := m.cfg.SecretByID(s)
				if secret != nil {
					requiredSecrets = append(requiredSecrets, secret)
				}
			}
		}
	}
	return requiredSecrets
}

func (m *Manager) fetchSecret(s *config.Secret) error {
	if m.server == nil {
		return errors.New("no server configured")
	}
	secret, err := m.server.Secret(s.SecretID)
	if err != nil {
		return err
	}
	value, ok := secret.Field(s.Field)
	if !ok {
		return fmt.Errorf("field %q does not exist", s.Field)
	}
	s.Value = value
	return nil
}

func (m *Manager) fetchRequiredSecrets(secrets []*config.Secret) int {
	for _, s := range secrets {
		if err := m.fetchSecret(s); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "forbidden") {
				log.Warn("Failed to fetch secret!", "id", s.SecretID, "field", s.Field, "err", err)
				if err := m.Authenticate(false, true); err != nil {
					log.Warn("Authentication failed", "err", err)
				}
				if err := m.fetchSecret(s); err == nil {
					continue
				}
			}
			log.Fatal("Failed to fetch secret!", "id", s.SecretID, "field", s.Field, "err", err)
		}
	}
	m.secrets = secrets
	return len(secrets)
}

func (m *Manager) secretByID(id string) *config.Secret {
	for _, s := range m.secrets {
		if strings.EqualFold(s.ID, id) {
			return s
		}
	}
	return nil
}
