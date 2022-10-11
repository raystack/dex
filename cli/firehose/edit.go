package firehose

import (
	"github.com/spf13/cobra"
)

func editCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Edit a firehose",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}
	return cmd
}
