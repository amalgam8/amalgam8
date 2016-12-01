package config

import (
	"os"

	cmds "github.com/amalgam8/amalgam8/cli/commands"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/urfave/cli"
)

// GlobalCommands .
func GlobalCommands(term terminal.UI) []cli.Command {
	if term == nil {
		term = terminal.NewUI(os.Stdin, os.Stdout)
	}

	return []cli.Command{
		cmds.NewServiceListCommand(term).GetMetadata(),
		cmds.NewRouteListCommand(term).GetMetadata(),
		cmds.NewActionListCommand(term).GetMetadata(),
		cmds.NewRuleCreateCommand(term).GetMetadata(),
		cmds.NewRuleGetCommand(term).GetMetadata(),
		cmds.NewRuleDeleteCommand(term).GetMetadata(),
		cmds.NewTrafficStartCommand(term).GetMetadata(),
		cmds.NewTrafficStepCommand(term).GetMetadata(),
		cmds.NewInfoCommand(term).GetMetadata(),
	}
}
