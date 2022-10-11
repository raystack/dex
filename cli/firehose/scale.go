package firehose

import (
	"github.com/spf13/cobra"
)

func scaleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scale",
		Short: "Scale a firehose",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}
	return cmd
}
