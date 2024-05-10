package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

type Config struct {
	SecretServer *SecretServer `mapstructure:"secret_server"`
	Secrets      []*Secret     `mapstructure:"secrets"`
	Groups       []*Group      `mapstructure:"groups"`
}

type SecretServer struct {
	URL string `mapstructure:"url"`
	TTL uint   `mapstructure:"ttl"`
}

type Secret struct {
	ID       string `mapstructure:"id"`
	Value    string `mapstructure:"-"`
	SecretID int    `mapstructure:"secret_id"`
	Field    string `mapstructure:"field"`
}

type Group struct {
	Name      string      `mapstructure:"name"`
	Selected  bool        `mapstructure:"selected"`
	Injectors []*Injector `mapstructure:"injectors"`
}

type Injector struct {
	Name     string            `mapstructure:"name"`
	Selected bool              `mapstructure:"selected"`
	Configs  []*InjectorConfig `mapstructure:"configs"`
}

type InjectorConfig struct {
	// Env secrets
	EnvKey    string `mapstructure:"env_key"`
	EnvSecret string `mapstructure:"env_secret"`
	// Secrets to stdout
	Stdout       bool   `mapstructure:"stdout"`
	StdoutSecret string `mapstructure:"stdout_secret"`
	// Templated secrets
	TmpFile        bool               `mapstructure:"tmp_file"`
	TmpFileSecrets []string           `mapstructure:"tmp_file_secrets"`
	TmpFileTmpl    string             `mapstructure:"tmp_file_tmpl"`
	Secrets        map[string]*Secret `mapstructure:"-"` // helper struct for templates
	TmpFileVar     string             `mapstructure:"tmp_file_var"`
	TmpFileSuffix  string             `mapstructure:"tmp_file_suffix"`
}

func init() {
	setDefaults()
}

func setDefaults() {
	//viper.SetDefault("secret_server.url", "https://my-secret-server.com")
	viper.SetDefault("secret_server.ttl", 7200)
}

func Load(path string) (cfg *Config, err error) {
	var loadedFrom string
	if path != "" {
		cfg, loadedFrom, err = loadCfg(path, true)
		if err != nil {
			return nil, err
		}
		log.Debug("Loaded config.", "path", loadedFrom, "flag", true)

		return cfg, nil
	}
	for _, f := range &[...]string{
		".esi.yml",
		"esi.yml",
		".esi.yaml",
		"esi.yaml",
	} {
		cfg, loadedFrom, err = loadCfg(f, false)
		if err != nil && os.IsNotExist(err) {
			err = nil
			continue
		} else if err != nil && errors.As(err, &viper.ConfigFileNotFoundError{}) {
			err = nil
			continue
		}
		log.Debug("Loaded config.", "path", loadedFrom, "flag", false)
		break
	}

	if cfg == nil {
		return cfg, viper.Unmarshal(&cfg)
	}

	return
}

func loadCfg(file string, explicit bool) (cfg *Config, loadedFrom string, err error) {
	if explicit {
		viper.SetConfigFile(file)
	} else {
		viper.SetConfigName(file)
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./")
		viper.AddConfigPath(filepath.Join(xdg.ConfigHome, "esi"))
		viper.AddConfigPath("/etc/esi/")
	}
	if err = viper.ReadInConfig(); err != nil {
		return
	}
	if err = viper.Unmarshal(&cfg); err != nil {
		return
	}
	loadedFrom = viper.ConfigFileUsed()
	return
}

func (c *Config) SecretByID(id string) *Secret {
	for _, s := range c.Secrets {
		if strings.EqualFold(s.ID, id) {
			return s
		}
	}
	log.Error("Failed to find secret by ID in config!", "id", id)
	return nil
}

func (c *Config) InjectorsByGroupName(name string) []*Injector {
	for _, g := range c.Groups {
		if strings.EqualFold(g.Name, name) {
			return g.Injectors
		}
	}
	log.Error("Failed to find injectors by group name!", "name", name)
	return nil
}

func (c *Config) InjectorByFQDN(fqdn string) *Injector {
	l := func(s string) string { return strings.ToLower(s) }
	for _, g := range c.Groups {
		if strings.HasPrefix(l(fqdn), l(g.Name)) {
			for _, i := range g.Injectors {
				if strings.EqualFold(fqdn, g.Name+"."+i.Name) {
					return i
				}
			}
		}
	}
	log.Error("Failed to find injector by fqdn", "fqdn", fqdn)
	return nil
}
