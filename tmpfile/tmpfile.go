package tmpfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jon4hz/esi/config"
)

type TmpFile struct {
	f *os.File
}

func New(injector *config.InjectorConfig, uid string) (*TmpFile, error) {
	tmpDir := os.TempDir()
	prefix := "esitmp-" + uid + "-"
	if err := cleanupOldFolder(tmpDir, prefix); err != nil {
		return nil, err
	}

	f, err := os.CreateTemp(tmpDir, prefix)
	if err != nil {
		return nil, err
	}

	if suffix := injector.TmpFileSuffix; suffix != "" {
		newName := f.Name() + suffix
		if err := f.Close(); err != nil {
			return nil, fmt.Errorf("failed to close tmpfile: %w", err)
		}
		if err := os.Rename(f.Name(), newName); err != nil {
			return nil, fmt.Errorf("failed to rename tmpfile: %w", err)
		}

		var err error
		f, err = os.OpenFile(newName, os.O_RDWR|os.O_TRUNC|os.O_EXCL, 0600)
		if err != nil {
			return nil, fmt.Errorf("failed to close tmpfile: %w", err)
		}
	}

	tmpl, err := template.New(f.Name()).Parse(injector.TmpFileTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	if err := tmpl.Execute(f, injector); err != nil {
		return nil, fmt.Errorf("failed to exec template: %w", err)
	}

	return &TmpFile{f}, nil
}

func cleanupOldFolder(tmpdir, prefix string) error {
	entries, err := os.ReadDir(tmpdir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), prefix) {
			if err := os.Remove(filepath.Join(tmpdir, e.Name())); err != nil {
				return fmt.Errorf("failed to cleanup old dir: %w", err)
			}
		}
	}
	return nil
}

func (t *TmpFile) Path() string {
	return t.f.Name()
}

func (t *TmpFile) Cleanup() (string, error) {
	t.f.Close() // nolint:errcheck
	return t.f.Name(), os.Remove(t.f.Name())
}
