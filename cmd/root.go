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

var rootCmdFlags struct {
	path     string
	debug    bool
	injector string
}

var rootCmd = &cobra.Command{
	Version:           version.Version,
	Use:               "esi",
	Short:             "fetch secrets from TSS and inject them to other processes in form of environment variables, config files or from stdin.",
	Args:              cobra.MinimumNArgs(1),
	CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	Run:               runRoot,
	Example:           `esi -- echo $MY_SECRET"`,
}

func init() {
	rootCmd.Flags().StringVarP(&rootCmdFlags.path, "config", "c", "", "path to the config file")
	rootCmd.Flags().StringVar(&rootCmdFlags.injector, "injector", "", fmt.Sprintf("fqdn of the injector (loads value from %s by default)", workspace.ESIWorkspaceFileName))
	rootCmd.Flags().BoolVar(&rootCmdFlags.debug, "debug", false, "enable debug logs")

	rootCmd.AddCommand(
		versionCmd,
		shellCmd,
		loginCmd,
	)
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func runRoot(cmd *cobra.Command, args []string) {
	if rootCmdFlags.debug {
		log.SetLevel(log.DebugLevel)
	}

	cfg, err := config.Load(rootCmdFlags.path)
	if err != nil {
		log.Fatal("Failed to load config", "err", err)
	}

	var inj *config.Injector
	if cmd.Flags().Lookup("injector").Changed {
		if inj = cfg.InjectorByFQDN(rootCmdFlags.injector); inj != nil {
			log.Debug("Loaded injector fqdn from injector flag.", "fqdn", rootCmdFlags.injector)
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

	if err := mgr.Run(false); err != nil {
		log.Fatal("Manager failed!", "err", err)
	}
}
