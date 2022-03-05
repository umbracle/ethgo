package commands

import (
	"github.com/mitchellh/cli"
	"github.com/umbracle/ethgo/cmd/version"
)

// VersionCommand is the command to show the version of the agent
type VersionCommand struct {
	UI cli.Ui
}

// Help implements the cli.Command interface
func (c *VersionCommand) Help() string {
	return `Usage: ethgo version

  Display the Ethgo version`
}

// Synopsis implements the cli.Command interface
func (c *VersionCommand) Synopsis() string {
	return "Display the Ethgo version"
}

// Run implements the cli.Command interface
func (c *VersionCommand) Run(args []string) int {
	c.UI.Output(version.GetVersion())
	return 0
}
