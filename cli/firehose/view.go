package firehose

import (
	"fmt"

	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/generated/client/operations"
)

func viewCommand() *cobra.Command {
	var project, name string

	cmd := &cobra.Command{
		Use:   "view",
		Short: "View a firehose",
		Long:  "Display information about a firehose",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		spinner := printer.Spin("")
		defer spinner.Stop()

		client := initClient()

		params := operations.GetFirehoseParams{
			FirehoseUrn: name,
			ProjectID:   project,
		}
		res, err := client.Operations.GetFirehose(&params)
		if err != nil {
			return err
		}

		firehose := res.Payload
		fmt.Println(firehose)

		return nil
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Firehose URN")
	cmd.MarkFlagRequired("namespace")

	cmd.Flags().StringVarP(&project, "project", "n", "", "Name of the project")
	cmd.MarkFlagRequired("project")

	return cmd
}
