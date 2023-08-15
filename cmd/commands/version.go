package commands

import (
	"github.com/mitchellh/cli"
	"github.com/yellomoon/web3tool/cmd/version"
)

// VersionCommand is the command to show the version of the agent
type VersionCommand struct {
	UI cli.Ui
}

// Help implements the cli.Command interface
func (c *VersionCommand) Help() string {
	return `Usage: web3tool version

  Display the Web3tool version`
}

// Synopsis implements the cli.Command interface
func (c *VersionCommand) Synopsis() string {
	return "Display the Web3tool version"
}

// Run implements the cli.Command interface
func (c *VersionCommand) Run(args []string) int {
	c.UI.Output(version.GetVersion())
	return 0
}
