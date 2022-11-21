package firehose

import (
	"fmt"

	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/generated/client/operations"
)

func scaleCommand(cfgLoader ConfigLoader) *cobra.Command {
	var replicas int

	cmd := &cobra.Command{
		Use:   "scale <project> <firehoseURN>",
		Short: "Scale number of instances of the firehose",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client := initClient(cfgLoader)

			replicasNum := float64(replicas)
			params := &operations.ScaleFirehoseParams{
				FirehoseUrn: args[1],
				ProjectID:   args[0],
				Body: operations.ScaleFirehoseBody{
					Replicas: &replicasNum,
				},
			}

			_, err := client.Operations.ScaleFirehose(params)
			if err != nil {
				return err
			}

			fmt.Println("Scale request accepted. Use view command to check status.")
			return nil
		},
	}

	cmd.Flags().IntVarP(&replicas, "replicas", "r", 1, "Number of replicas to run")
	return cmd
}
