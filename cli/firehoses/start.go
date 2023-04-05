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

func startCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start <project> <firehoseURN>",
		Short: "Start the firehose if it's currently stopped.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error { //nolint:dupl
			spinner := printer.Spin("")
			defer spinner.Stop()

			params := &operations.StartFirehoseParams{
				FirehoseUrn: args[0],
				Body:        struct{}{},
			}

			dexAPI := cdk.NewClient(cmd)
			modifiedFirehose, err := dexAPI.Operations.StartFirehose(params)
			if err != nil {
				return err
			}
			spinner.Stop()

			return cdk.Display(cmd, modifiedFirehose, func(w io.Writer, v interface{}) error {
				_, err := fmt.Fprintln(w, "Start request accepted. Use view command to check status.")
				return err
			})
		},
	}
	return cmd
}
