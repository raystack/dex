package firehose

import (
	"fmt"

	"github.com/spf13/cobra"
)

func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List a firehose",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Println("Creating firehose...")
		return nil
	}
	return cmd
}
