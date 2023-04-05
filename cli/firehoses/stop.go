package firehoses

import (
	"fmt"
	"io"

	"github.com/goto/salt/printer"
	"github.com/spf13/cobra"

	"github.com/goto/dex/cli/cdk"
	"github.com/goto/dex/generated/client/operations"
)

func stopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop <firehoseURN>",
		Short: "Stop the firehose if it's currently running.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			dexAPI := cdk.NewClient(cmd)
			params := &operations.StopFirehoseParams{
				FirehoseUrn: args[0],
				Body:        struct{}{},
			}

			modifiedFirehose, err := dexAPI.Operations.StopFirehose(params)
			if err != nil {
				return err
			}

			return cdk.Display(cmd, modifiedFirehose, func(w io.Writer, v interface{}) error {
				_, err := fmt.Fprintln(w, "Stop request accepted. Use view command to check status.")
				return err
			})
		},
	}
	return cmd
}
