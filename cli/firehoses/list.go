package firehoses

import (
	"fmt"
	"io"

	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/cli/cdk"
	"github.com/odpf/dex/generated/client/operations"
	"github.com/odpf/dex/generated/models"
	"github.com/odpf/dex/pkg/errors"
)

func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project>",
		Short: "List firehoses in the given project.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			firehoses, err := listFirehoses(cmd, args[0])
			if err != nil {
				return errors.Errorf("failed to list: %s", err)
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

func listFirehoses(cmd *cobra.Command, prjSlug string) ([]*models.Firehose, error) {
	spinner := printer.Spin(fmt.Sprintf("Fetching firehoses in project '%s'", prjSlug))
	defer spinner.Stop()

	params := operations.ListFirehosesParams{
		ProjectSlug: prjSlug,
	}

	dexAPI := initClient(cmd)
	res, err := dexAPI.Operations.ListFirehoses(&params)
	if err != nil {
		return nil, err
	}
	return res.GetPayload().Items, nil
}
