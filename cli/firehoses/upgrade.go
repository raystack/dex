//nolint:dupl
package firehoses

import (
	"fmt"
	"io"

	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/cli/cdk"
	"github.com/odpf/dex/generated/client/operations"
)

func upgradeCommand() *cobra.Command {
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
