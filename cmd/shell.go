package cmd

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/jon4hz/esi/config"
	"github.com/jon4hz/esi/manager"
	"github.com/jon4hz/esi/version"
	"github.com/jon4hz/esi/workspace"
	"github.com/spf13/cobra"
)

var shellCmdFlags struct {
	path     string
	injector string
	debug    bool
}

var shellCmd = &cobra.Command{
	Version: version.Version,
	Use:     "shell",
	Short:   "Spawn the command in a subshell",
	Args:    cobra.MinimumNArgs(1),
	Run:     runShell,
	Example: `esi shell -- \"env | grep MY_SECRET || echo could not find my secret."`,
}

func init() {
	shellCmd.Flags().StringVarP(&shellCmdFlags.path, "config", "c", "", "path to the config file")
	shellCmd.Flags().StringVar(&shellCmdFlags.injector, "injector", "", fmt.Sprintf("fqdn of the injector (loads value from %s by default)", workspace.ESIWorkspaceFileName))
	shellCmd.Flags().BoolVar(&shellCmdFlags.debug, "debug", false, "enable debug logs")
}

func runShell(cmd *cobra.Command, args []string) {
	if shellCmdFlags.debug {
		log.SetLevel(log.DebugLevel)
	}

	cfg, err := config.Load(shellCmdFlags.path)
	if err != nil {
		log.Fatal("Failed to load config", "err", err)
	}

	var inj *config.Injector
	if cmd.Flags().Lookup("injector").Changed {
		if inj = cfg.InjectorByFQDN(shellCmdFlags.injector); inj != nil {
			log.Debug("Loaded injector fqdn from injector flag.", "fqdn", shellCmdFlags.injector)
		}
	} else {
		wscfg := workspace.New()
		if wscfg != nil {
			if inj = cfg.InjectorByFQDN(wscfg.Injector); inj != nil {
				log.Debug("Loaded injector fqdn from workspace file", "fqdn", wscfg.Injector)
			}
		}
	}

	mgr, err := manager.New(cfg, args, inj)
	if err != nil {
		log.Fatal("Failed to create manager", "err", err)
	}

	if err := mgr.Run(true); err != nil {
		log.Fatal("Manager failed!", "err", err)
	}
}
