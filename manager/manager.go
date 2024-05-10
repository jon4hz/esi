package manager

import (
	"fmt"
	"os"
	"os/user"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/jon4hz/esi/config"
	"github.com/jon4hz/esi/forms"
	"github.com/jon4hz/esi/keyring"
	"github.com/jon4hz/tss-sdk-go/v2/server"
)

const (
	passwordID = "esi:password"
	tokenID    = "esi:token"
)

type Manager struct {
	cfg        *config.Config
	server     *server.Server
	sKeyring   *keyring.Keyring
	uKeyring   *keyring.Keyring
	args       []string
	env        []string
	secrets    []*config.Secret
	injector   *config.Injector
	currentUID string
	cleanup    func()
	cleanupMu  sync.Mutex
	cleanDone  bool
}

func New(cfg *config.Config, args []string, inj *config.Injector) (*Manager, error) {
	m := Manager{
		cfg:      cfg,
		args:     args,
		env:      os.Environ(),
		injector: inj,
	}

	user, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	m.currentUID = user.Uid

	m.sKeyring, err = keyring.New(keyring.WithKeyringType(keyring.SessionKeyring))
	if err != nil {
		return nil, err
	}

	m.uKeyring, err = keyring.New(keyring.WithKeyringType(keyring.UserKeyring))
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *Manager) Run(subshell bool) error {
	if err := m.Authenticate(false, false); err != nil {
		return err
	}

	if m.injector == nil {
		var group *config.Group
		if err := forms.GroupSelectForm(m.cfg.Groups, &group).Run(); err != nil {
			return err
		}

		var injector *config.Injector
		if err := forms.InjectorSelectForm(m.cfg.InjectorsByGroupName(group.Name), &injector).Run(); err != nil {
			return err
		}
		m.injector = injector
	}

	const maxRetries = 3
	var found int
	requiredSecrets := m.requiredSecrets(m.injector)
	if len(requiredSecrets) == 0 {
		log.Fatal("No required secrets found. Aborting...")
	}

	for i := 0; i < maxRetries; i++ {
		found = m.fetchRequiredSecrets(requiredSecrets)
		if found != 0 {
			break
		}
	}
	if found == 0 {
		log.Fatal("Unable to fetch any secrets!", "err", "max retries exceeded")
	}

	cleaners := m.deployTmpFiles(m.injector.Configs)
	m.cleanup = func() {
		m.cleanupMu.Lock()
		if m.cleanDone {
			m.cleanupMu.Unlock()
			return
		}
		for _, c := range cleaners {
			if path, err := c(); err != nil {
				log.Warn("Cleanup failed. Please delete the tmpfile manually!", "path", path)
			} else {
				log.Info("Cleanup successful!", "path", path)
			}
		}
		m.cleanDone = true
		m.cleanupMu.Unlock()
	}

	defer func() {
		m.cleanup()
	}()

	m.printSecrets(m.injector.Configs)

	m.setEnvVars(m.injector.Configs)

	if subshell {
		if _, err := m.execSubshell(m.args, m.env); err != nil {
			return err
		}
	} else {
		if err := m.executeSingleCommandWithEnvs(m.args); err != nil {
			return err
		}
	}
	return nil
}
