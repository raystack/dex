//nolint:dupl
package firehose

import (
	"fmt"

	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/generated/client/operations"
)

func upgradeCommand(cfgLoader ConfigLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade <project> <firehoseURN>",
		Short: "Upgrade the firehose to the latest version supported",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			params := &operations.UpgradeFirehoseParams{
				FirehoseUrn: args[1],
				ProjectSlug: args[0],
				Body:        struct{}{},
			}

			client := initClient(cfgLoader)
			_, err := client.Operations.UpgradeFirehose(params)
			if err != nil {
				return err
			}

			fmt.Println("Upgrade request accepted. Use view command to check status.")
			return nil
		},
	}
	return cmd
}
