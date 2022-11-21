package firehose

import (
	"fmt"

	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/generated/client/operations"
)

func viewCommand(cfgLoader ConfigLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view <project> <name>",
		Short: "View a firehose",
		Long:  "Display information about a firehose",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client := initClient(cfgLoader)

			params := operations.GetFirehoseParams{
				ProjectID:   args[0],
				FirehoseUrn: args[1],
			}
			res, err := client.Operations.GetFirehose(&params)
			if err != nil {
				return err
			}
			firehose := res.Payload

			fmt.Println(firehose)
			return nil
		},
	}

	return cmd
}
