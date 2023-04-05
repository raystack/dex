package firehoses

import (
	"fmt"
	"io"

	"github.com/goto/salt/printer"
	"github.com/spf13/cobra"

	"github.com/goto/dex/cli/cdk"
	"github.com/goto/dex/generated/client/operations"
	"github.com/goto/dex/generated/models"
	"github.com/goto/dex/pkg/errors"
)

func scaleCommand() *cobra.Command {
	var replicas int

	cmd := &cobra.Command{
		Use:   "scale <firehoseURN>",
		Short: "Scale number of instances of the firehose",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modifiedFirehose, err := scaleFirehose(cmd, args[0], replicas)
			if err != nil {
				return errors.Errorf("scale operation failed: %s", err)
			}

			return cdk.Display(cmd, modifiedFirehose, func(w io.Writer, v interface{}) error {
				_, err := fmt.Fprintln(w, "Scale request accepted. Use view command to check status.")
				return err
			})
		},
	}

	cmd.Flags().IntVarP(&replicas, "replicas", "r", 1, "Number of replicas to run")
	return cmd
}

func scaleFirehose(cmd *cobra.Command, urn string, replicas int) (*models.Firehose, error) {
	spinner := printer.Spin("")
	defer spinner.Stop()

	replicasNum := float64(replicas)
	params := &operations.ScaleFirehoseParams{
		FirehoseUrn: urn,
		Body: operations.ScaleFirehoseBody{
			Replicas: &replicasNum,
		},
	}

	dexAPI := cdk.NewClient(cmd)
	resp, err := dexAPI.Operations.ScaleFirehose(params)
	if err != nil {
		return nil, err
	}
	return resp.GetPayload(), nil
}
