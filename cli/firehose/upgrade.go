package firehose

import (
	"github.com/spf13/cobra"
)

func upgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade a firehose",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}
	return cmd
}
