package cmd

import (
	"github.com/spf13/cobra"
)

type ingressTraefikConverterCommands struct {
	commands []*cobra.Command
}

// SetIngressTraefikConverterCommands helps in gathering all the subcommands so that it can be used while registering it with main command.
func SetIngressTraefikConverterCommands() *cobra.Command {
	return getIngressTraefikConverterCommands()
}

// Add an entry in below function to register new command.
func getIngressTraefikConverterCommands() *cobra.Command {
	command := new(ingressTraefikConverterCommands)
	command.commands = append(command.commands, getImportCommand())
	command.commands = append(command.commands, getVersionCommand())

	return command.prepareCommands()
}

func (c *ingressTraefikConverterCommands) prepareCommands() *cobra.Command {
	rootCmd := getRootCommand()
	for _, command := range c.commands {
		rootCmd.AddCommand(command)
	}

	registerCommonFlags(rootCmd)

	return rootCmd
}

func getUsageTemplate() string {
	return `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}{{printf "\n" }}
Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}{{printf "\n" }}
Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{printf "\n"}}
Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}{{printf "\n"}}
Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}{{printf "\n"}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}{{printf "\n"}}
Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}
{{if .HasAvailableSubCommands}}{{printf "\n"}}
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
{{printf "\n"}}`
}
