package commands

import (
	"os"

	"github.com/cloudwalk/ethgo/ens"
	"github.com/mitchellh/cli"
	flag "github.com/spf13/pflag"
)

func defaultJsonRPCProvider() string {
	if provider := os.Getenv("JSONRPC_PROVIDER"); provider != "" {
		return provider
	}
	return "http://localhost:8545"
}

// EnsResolveCommand is the command to resolve an ens name
type EnsResolveCommand struct {
	UI cli.Ui

	provider string
}

// Help implements the cli.Command interface
func (c *EnsResolveCommand) Help() string {
	return `Usage: ethgo ens resolve <name>

  Resolve an ens name
` + c.Flags().FlagUsages()
}

// Synopsis implements the cli.Command interface
func (c *EnsResolveCommand) Synopsis() string {
	return "Resolve an ens name"
}

func (c *EnsResolveCommand) Flags() *flag.FlagSet {
	flags := flag.NewFlagSet("ens resolve", flag.PanicOnError)
	flags.StringVar(&c.provider, "provider", defaultJsonRPCProvider(), "")

	return flags
}

// Run implements the cli.Command interface
func (c *EnsResolveCommand) Run(args []string) int {
	flags := c.Flags()
	if err := flags.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	args = flags.Args()
	if len(args) != 1 {
		c.UI.Error("one argument <name> expected")
		return 1
	}

	e, err := ens.NewENS(ens.WithAddress(c.provider))
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	addr, err := e.Resolve(args[0])
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	c.UI.Output(addr.String())
	return 0
}
