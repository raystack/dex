//nolint:dupl
package firehoses

import (
	"fmt"
	"io"

	"github.com/goto/salt/printer"
	"github.com/spf13/cobra"

	"github.com/goto/dex/cli/cdk"
	"github.com/goto/dex/generated/client/operations"
)

func upgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade <firehoseURN>",
		Short: "Upgrade the firehose to the latest version supported",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			params := &operations.UpgradeFirehoseParams{
				FirehoseUrn: args[1],
				Body:        struct{}{},
			}

			dexAPI := cdk.NewClient(cmd)
			modifiedFirehose, err := dexAPI.Operations.UpgradeFirehose(params)
			if err != nil {
				return err
			}

			return cdk.Display(cmd, modifiedFirehose, func(w io.Writer, v interface{}) error {
				_, err := fmt.Fprintln(w, "Upgrade request accepted. Use view command to check status.")
				return err
			})
		},
	}
	return cmd
}
