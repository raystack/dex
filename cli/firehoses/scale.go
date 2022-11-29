package firehoses

import (
	"fmt"
	"io"

	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/cli/cdk"
	"github.com/odpf/dex/generated/client/operations"
)

func scaleCommand() *cobra.Command {
	var replicas int

	cmd := &cobra.Command{
		Use:   "scale <project> <firehoseURN>",
		Short: "Scale number of instances of the firehose",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client := initClient(cmd)

			replicasNum := float64(replicas)
			params := &operations.ScaleFirehoseParams{
				FirehoseUrn: args[1],
				ProjectSlug: args[0],
				Body: operations.ScaleFirehoseBody{
					Replicas: &replicasNum,
				},
			}

			modifiedFirehose, err := client.Operations.ScaleFirehose(params)
			if err != nil {
				return err
			}
			spinner.Stop()

			return cdk.Display(cmd, modifiedFirehose, func(w io.Writer, v interface{}) error {
				_, err := fmt.Fprintln(w, "Scale request accepted. Use view command to check status.")
				return err
			})
		},
	}

	cmd.Flags().IntVarP(&replicas, "replicas", "r", 1, "Number of replicas to run")
	return cmd
}
