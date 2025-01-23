package root

import (
	"fmt"
	"os"

	"github.com/apex-fusion/nexus/command/backup"
	"github.com/apex-fusion/nexus/command/genesis"
	"github.com/apex-fusion/nexus/command/helper"
	"github.com/apex-fusion/nexus/command/ibft"
	"github.com/apex-fusion/nexus/command/license"
	"github.com/apex-fusion/nexus/command/monitor"
	"github.com/apex-fusion/nexus/command/peers"
	"github.com/apex-fusion/nexus/command/secrets"
	"github.com/apex-fusion/nexus/command/server"
	"github.com/apex-fusion/nexus/command/status"
	"github.com/apex-fusion/nexus/command/version"
	"github.com/apex-fusion/nexus/command/whitelist"
	"github.com/spf13/cobra"
)

type RootCommand struct {
	baseCmd *cobra.Command
}

func NewRootCommand() *RootCommand {
	rootCommand := &RootCommand{
		baseCmd: &cobra.Command{
			Short: "Nexus is a higly-performant EVM runtime client and IBFT-2.0 consensus client",
		},
	}

	helper.RegisterJSONOutputFlag(rootCommand.baseCmd)

	rootCommand.registerSubCommands()

	return rootCommand
}

func (rc *RootCommand) registerSubCommands() {
	rc.baseCmd.AddCommand(
		version.GetCommand(),
		status.GetCommand(),
		secrets.GetCommand(),
		peers.GetCommand(),
		monitor.GetCommand(),
		ibft.GetCommand(),
		backup.GetCommand(),
		genesis.GetCommand(),
		server.GetCommand(),
		whitelist.GetCommand(),
		license.GetCommand(),
	)
}

func (rc *RootCommand) Execute() {
	if err := rc.baseCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}
