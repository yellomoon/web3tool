package commands

import (
	"os"

	"github.com/mitchellh/cli"
	flag "github.com/spf13/pflag"
)

func Commands() map[string]cli.CommandFactory {
	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	baseCommand := &baseCommand{
		UI: ui,
	}

	return map[string]cli.CommandFactory{
		"abigen": func() (cli.Command, error) {
			return &AbigenCommand{
				baseCommand: baseCommand,
			}, nil
		},
		"4byte": func() (cli.Command, error) {
			return &FourByteCommand{
				UI: ui,
			}, nil
		},
		"ens": func() (cli.Command, error) {
			return &EnsCommand{
				UI: ui,
			}, nil
		},
		"ens resolve": func() (cli.Command, error) {
			return &EnsResolveCommand{
				UI: ui,
			}, nil
		},
		"version": func() (cli.Command, error) {
			return &VersionCommand{
				UI: ui,
			}, nil
		},
	}
}

type baseCommand struct {
	UI cli.Ui
}

func (b *baseCommand) Flags(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, 0)
	return flags
}
