package commands

import (
	flag "github.com/spf13/pflag"
	"github.com/umbracle/ethgo/cmd/abigen"
)

// VersionCommand is the command to show the version of the agent
type AbigenCommand struct {
	*baseCommand

	source string
	pckg   string
	output string
	name   string
}

// Help implements the cli.Command interface
func (c *AbigenCommand) Help() string {
	return `Usage: ethgo abigen
	
  Compute the abigen.
` + c.Flags().FlagUsages()
}

// Synopsis implements the cli.Command interface
func (c *AbigenCommand) Synopsis() string {
	return "Compute the abigen"
}

func (c *AbigenCommand) Flags() *flag.FlagSet {
	flags := c.baseCommand.Flags("abigen")

	flags.StringVar(&c.source, "source", "", "Source data")
	flags.StringVar(&c.pckg, "package", "main", "Name of the package")
	flags.StringVar(&c.output, "output", ".", "Output directory")
	flags.StringVar(&c.name, "name", "", "name of the contract")

	return flags
}

// Run implements the cli.Command interface
func (c *AbigenCommand) Run(args []string) int {
	flags := c.Flags()
	if err := flags.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	if err := abigen.Parse(c.source, c.pckg, c.output, c.name); err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	return 0
}
