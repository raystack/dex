package firehose

import (
	"github.com/spf13/cobra"
)

func viewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View a firehose",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}
	return cmd
}
