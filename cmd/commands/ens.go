package commands

import (
	"github.com/mitchellh/cli"
)

// EnsCommand is the group command for ens
type EnsCommand struct {
	UI cli.Ui
}

// Help implements the cli.Command interface
func (c *EnsCommand) Help() string {
	return `Usage: ethgo ens

  Interact with ens`
}

// Synopsis implements the cli.Command interface
func (c *EnsCommand) Synopsis() string {
	return "Interact with ens"
}

// Run implements the cli.Command interface
func (c *EnsCommand) Run(args []string) int {
	return 0
}
