package firehoses

import (
	"fmt"
	"io"
	"log"

	"github.com/goto/salt/printer"
	"github.com/goto/salt/term"
	"github.com/spf13/cobra"

	"github.com/goto/dex/cli/cdk"
	"github.com/goto/dex/generated/client/operations"
	"github.com/goto/dex/generated/models"
)

func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project>",
		Short: "List firehoses in the given project.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params := operations.ListFirehosesParams{
				Project: args[0],
			}

			firehoses, err := listFirehoses(cmd, params)
			if err != nil {
				log.Fatalf("failed to list: %s", err)
			}

			return cdk.Display(cmd, firehoses, func(w io.Writer, v interface{}) error {
				report := [][]string{{term.Bold("URN"), term.Bold("NAME")}}
				for _, f := range firehoses {
					report = append(report, []string{f.Urn, f.Name})
				}
				_, _ = fmt.Fprintf(w, "Showing %d firehoses\n", len(firehoses))
				printer.Table(w, report)
				return nil
			})
		},
	}

	return cmd
}

func listFirehoses(cmd *cobra.Command, params operations.ListFirehosesParams) ([]*models.Firehose, error) {
	spinner := printer.Spin(fmt.Sprintf("Fetching firehoses in project '%s'", params.Project))
	defer spinner.Stop()

	dexAPI := cdk.NewClient(cmd)
	res, err := dexAPI.Operations.ListFirehoses(&params)
	if err != nil {
		return nil, err
	}
	return res.GetPayload().Items, nil
}
