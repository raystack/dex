package firehoses

import (
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "firehose <command>",
		Aliases: []string{"s"},
		Short:   "Firehose management commands.",
		Long:    "You can create/manage/view firehoses using this command.",
		Example: heredoc.Doc(`
			$ dex firehose list project-x
			$ dex firehose create -c ./config.yaml
		`),
		Annotations: map[string]string{
			"group": "core",
		},
	}

	cmd.AddCommand(
		viewCommand(),
		listCommand(),
		applyCommand(),
		scaleCommand(),
		startCommand(),
		stopCommand(),
		logsCommand(),
		upgradeCommand(),
		resetOffsetCommand(),
	)

	cmd.PersistentFlags().DurationP("timeout", "T", 10*time.Second, "Timeout for the operation")
	return cmd
}
