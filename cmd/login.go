package cmd

import (
	"github.com/charmbracelet/log"
	"github.com/jon4hz/esi/config"
	"github.com/jon4hz/esi/manager"
	"github.com/spf13/cobra"
)

var loginCmdFlags struct {
	path  string
	debug bool
	force bool
}

var loginCmd = &cobra.Command{
	Use:  "login",
	Args: cobra.NoArgs,
	Run:  runLogin,
}

func init() {
	loginCmd.Flags().StringVarP(&loginCmdFlags.path, "config", "c", "", "path to the config file")
	loginCmd.Flags().BoolVar(&loginCmdFlags.debug, "debug", false, "enable debug logs")
	loginCmd.Flags().BoolVarP(&loginCmdFlags.force, "force", "f", false, "force new credentials")
}

func runLogin(_ *cobra.Command, _ []string) {
	if loginCmdFlags.debug {
		log.SetLevel(log.DebugLevel)
	}

	cfg, err := config.Load(loginCmdFlags.path)
	if err != nil {
		log.Fatal("Failed to load config", "err", err)
	}

	mgr, err := manager.New(cfg, nil, nil)
	if err != nil {
		log.Fatal("Failed to create manager", "err", err)
	}

	if err := mgr.Authenticate(loginCmdFlags.force, loginCmdFlags.force); err != nil {
		log.Fatal("Authentication failed!", "err", err)
	}
}
