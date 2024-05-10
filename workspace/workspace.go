package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
)

const ESIWorkspaceFileName = ".esi-workspace.yml"

var ErrFileNotFound = errors.New(fmt.Sprintf("file not found: %s", ESIWorkspaceFileName))

type Workspace struct {
	Injector string `yaml:"injector"`
}

func New() *Workspace {
	w, err := load()
	if err != nil {
		log.Error("Failed to load workspace config", "err", err)
		return nil
	}
	return w
}

func load() (*Workspace, error) {
	path, err := findWorkspaceConfigFile()
	if err != nil {
		if errors.Is(err, ErrFileNotFound) {
			return nil, nil
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var w Workspace
	if err := yaml.Unmarshal(data, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

// findWorkspaceConfigFile searches for a .esi-workspace.yml file in the current directory and all parent directories.
func findWorkspaceConfigFile() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		path := filepath.Join(dir, ESIWorkspaceFileName)
		_, err := os.Stat(path)
		if err == nil {
			// file found
			log.Debug("Found workspace config!", "path", path)

			return path, nil
		}

		// check if we have reached the root directory
		if dir == filepath.Dir(dir) {
			break
		}

		// move up one directory
		dir = filepath.Dir(dir)
	}

	// file not found
	return "", ErrFileNotFound
}
