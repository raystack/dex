package firehose

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func FirehoseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "firehose <command>",
		Aliases: []string{"s"},
		Short:   "Server management",
		Long:    "Server management commands.",
		Example: heredoc.Doc(`
			$ dex firehose list
			$ dex firehose create -c ./config.yaml
		`),
		Annotations: map[string]string{
			"group": "core",
		},
	}

	cmd.AddCommand(
		createCommand(),
		listCommand(),
		viewCommand(),
		stopCommand(),
		applyCommand(),
		editCommand(),
		stopCommand(),
		upgradeCommand(),
		scaleCommand(),
	)
	return cmd
}
