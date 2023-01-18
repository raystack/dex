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

func startCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start <project> <firehoseURN>",
		Short: "Start the firehose if it's currently stopped.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error { //nolint:dupl
			spinner := printer.Spin("")
			defer spinner.Stop()

			params := &operations.StartFirehoseParams{
				FirehoseUrn: args[1],
				ProjectSlug: args[0],
				Body:        struct{}{},
			}

			dexAPI := initClient(cmd)
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
