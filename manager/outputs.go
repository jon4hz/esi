package manager

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/jon4hz/esi/config"
	"github.com/jon4hz/esi/tmpfile"
)

func (m *Manager) deployTmpFiles(injectors []*config.InjectorConfig) []func() (string, error) {
	var cleaners []func() (string, error)
	for _, inj := range injectors {
		if !inj.TmpFile {
			continue
		}

		inj.Secrets = func() map[string]*config.Secret {
			secrets := make(map[string]*config.Secret)
			for _, s := range inj.TmpFileSecrets {
				if secretByID := m.secretByID(s); secretByID != nil {
					secrets[s] = secretByID
				}
			}
			return secrets
		}()

		tf, err := tmpfile.New(inj, m.currentUID)
		if err != nil {
			log.Warn("Failed to create tmpfile", "err", err)
		} else {
			log.Debug("Created tmpfile", "path", tf.Path(), "var", inj.TmpFileVar)
			if inj.TmpFileVar != "" {
				m.addEnv(inj.TmpFileVar, tf.Path())
			}
			cleaners = append(cleaners, tf.Cleanup)
		}
	}
	return cleaners
}

func (m *Manager) printSecrets(injectors []*config.InjectorConfig) {
	var b strings.Builder
	for i, inj := range injectors {
		secret := m.secretByID(inj.StdoutSecret)
		if secret == nil {
			log.Debug("Unable to find secret by ID!", "id", inj.StdoutSecret)
			continue
		}
		if inj.Stdout && secret.Value != "" {
			b.WriteString(secret.Value)
			if i < len(injectors)-1 {
				b.WriteString("\n\n")
			}
		}
	}
	fmt.Fprint(os.Stdout, b.String())
}

func (m *Manager) setEnvVars(injectors []*config.InjectorConfig) {
	for _, inj := range injectors {
		secret := m.secretByID(inj.EnvSecret)
		if secret == nil {
			log.Debug("Unable to find secret by ID!", "id", inj.EnvSecret)
			continue
		}
		if inj.EnvKey != "" && secret.Value != "" {
			m.addEnv(inj.EnvKey, secret.Value)
		}
	}
}
