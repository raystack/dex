package firehoses

import (
	"github.com/goto/salt/printer"
	"github.com/spf13/cobra"

	"github.com/goto/dex/cli/cdk"
	"github.com/goto/dex/generated/client/operations"
	"github.com/goto/dex/generated/models"
)

func viewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "view <name>",
		Short:   "View a firehose",
		Long:    "Display information about a firehose",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"get", "show"},
		RunE: func(cmd *cobra.Command, args []string) error {
			firehose, err := getFirehose(cmd, args[0])
			if err != nil {
				return err
			}
			return cdk.Display(cmd, firehose, cdk.YAMLFormat)
		},
	}

	return cmd
}

func getFirehose(cmd *cobra.Command, firehoseID string) (*models.Firehose, error) {
	sp := printer.Spin("Fetching firehose...")
	defer sp.Stop()

	params := &operations.GetFirehoseParams{
		FirehoseUrn: firehoseID,
	}

	cl := cdk.NewClient(cmd)
	res, err := cl.Operations.GetFirehose(params)
	if err != nil {
		return nil, err
	}
	return res.GetPayload(), nil
}
