package firehose

import (
	"fmt"
	"os"

	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/generated/client/operations"
)

func listCommand(cfgLoader ConfigLoader) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list <project>",
		Short: "List firehoses in the given project.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client := initClient(cfgLoader)

			params := operations.ListFirehosesParams{
				ProjectSlug: args[0],
			}
			res, err := client.Operations.ListFirehoses(&params)
			if err != nil {
				return err
			}
			firehoses := res.Payload.Items

			report := [][]string{
				{term.Bold("INDEX"), term.Bold("NAME")},
			}
			for idx, f := range firehoses {
				report = append(report, []string{term.Greenf("#%d", idx+1), f.Name})
			}
			spinner.Stop()

			fmt.Printf("\nShowing %d firehoses", len(firehoses))
			printer.Table(os.Stdout, report)
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "L", 30, "Maximum number of firehoses to fetch (default 30)")
	return cmd
}
