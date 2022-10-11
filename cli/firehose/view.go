package firehose

import (
	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"
)

func viewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View a firehose",
		Long:  "Display information about a firehose",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		spinner := printer.Spin("")
		defer spinner.Stop()

		return nil
	}
	return cmd
}
