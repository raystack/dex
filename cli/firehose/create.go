package firehose

import (
	"github.com/spf13/cobra"
)

func createCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a firehose",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "./config.yaml", "Config file path")

	return cmd
}
