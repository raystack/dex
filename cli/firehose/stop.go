package firehose

import (
	"github.com/spf13/cobra"
)

func stopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop a running firehose",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}
	return cmd
}
