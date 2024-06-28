package commands

import (
	"github.com/mitchellh/cli"
	fourbyte "github.com/Ethernal-Tech/ethgo/4byte"
)

// FourByteCommand is the command to resolve 4byte actions
type FourByteCommand struct {
	UI cli.Ui
}

// Help implements the cli.Command interface
func (c *FourByteCommand) Help() string {
	return `Usage: ethgo 4byte [signature]

  Resolve a 4byte signature`
}

// Synopsis implements the cli.Command interface
func (c *FourByteCommand) Synopsis() string {
	return "Resolve a 4byte signature"
}

// Run implements the cli.Command interface
func (c *FourByteCommand) Run(args []string) int {
	if len(args) == 0 {
		c.UI.Output("No arguments provided")
		return 1
	}

	found, err := fourbyte.Resolve(args[0])
	if err != nil {
		c.UI.Output(err.Error())
		return 1
	}
	if found == "" {
		c.UI.Output("Resolve not found")
		return 1
	}

	c.UI.Output(found)
	return 0
}
