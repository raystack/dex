package firehose

import (
	"github.com/spf13/cobra"
)

func applyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a firehose",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}
	return cmd
}
