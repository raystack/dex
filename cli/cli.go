package cli

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/salt/cmdx"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/cli/auth"
	"github.com/odpf/dex/cli/config"
	"github.com/odpf/dex/cli/firehose"
	"github.com/odpf/dex/cli/server"
	"github.com/odpf/dex/pkg/version"
)

var envHelp = map[string]string{
	"short": "List of supported environment variables",
	"long": heredoc.Doc(`
			ODPF_CONFIG_DIR: the directory where dex will store configuration files. Default:
			"$XDG_CONFIG_HOME/odpf" or "$HOME/.config/odpf".

			NO_COLOR: set to any value to avoid printing ANSI escape sequences for color output.

			CLICOLOR: set to "0" to disable printing ANSI colors in output.
		`),
}

// New root command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "dex <command> <subcommand> [flags]",
		Short:         "Data experience console",
		Long:          "Data experience console.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Annotations: map[string]string{
			"group": "core",
			"help:learn": heredoc.Doc(`
				Use 'dex <command> --help' for info about a command.
				Read the manual at https://odpf.github.io/dex/
			`),
			"help:feedback": heredoc.Doc(`
				Open an issue here https://github.com/odpf/dex/issues
			`),
		},
	}

	cmd.AddCommand(
		versionCmd(),
		auth.LoginCommand(),
		config.Commands(),
		server.Commands(),
		firehose.Commands(),
	)

	// Help topics.
	cmdx.SetHelp(cmd)
	cmd.AddCommand(
		cmdx.SetCompletionCmd("dex"),
		cmdx.SetHelpTopicCmd("environment", envHelp),
		cmdx.SetRefCmd(cmd),
	)

	cmdx.SetClientHook(cmd, func(cmd *cobra.Command) {
		// client config.
		cmd.PersistentFlags().StringP("host", "h", "", "Server host address")
	})

	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return version.Print()
		},
	}
}
