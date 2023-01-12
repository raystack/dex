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
	var sinkType, group, status string
	cmd := &cobra.Command{
		Use:   "list <project>",
		Short: "List firehoses in the given project.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params := operations.ListFirehosesParams{
				Group:       strPtrIfNonEmpty(group),
				Status:      strPtrIfNonEmpty(status),
				SinkType:    strPtrIfNonEmpty(sinkType),
				ProjectSlug: args[0],
			}

			firehoses, err := listFirehoses(cmd, params)
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

	flags := cmd.Flags()
	flags.StringVarP(&group, "group", "G", "", "Consumer group to filter by")
	flags.StringVarP(&sinkType, "sink-type", "S", "", "Sink type to filter")
	flags.StringVarP(&status, "status", "s", "", "Status of the firehose deployment")
	return cmd
}

func listFirehoses(cmd *cobra.Command, params operations.ListFirehosesParams) ([]*models.Firehose, error) {
	spinner := printer.Spin(fmt.Sprintf("Fetching firehoses in project '%s'", params.ProjectSlug))
	defer spinner.Stop()

	dexAPI := initClient(cmd)
	res, err := dexAPI.Operations.ListFirehoses(&params)
	if err != nil {
		return nil, err
	}
	return res.GetPayload().Items, nil
}

func strPtrIfNonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
