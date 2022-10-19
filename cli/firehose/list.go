package firehose

import (
	"fmt"
	"os"

	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/generated/client/operations"
)

func listCommand() *cobra.Command {
	var limit int
	var project string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List firehoses",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		spinner := printer.Spin("")
		defer spinner.Stop()

		client := initClient()

		params := operations.ListFirehosesParams{
			ProjectID: project,
		}
		res, err := client.Operations.ListFirehoses(&params)
		if err != nil {
			return err
		}

		firehoses := res.Payload.Items

		report := [][]string{}

		index := 1
		report = append(report, []string{
			term.Bold("INDEX"),
			term.Bold("NAME"),
		})
		for _, f := range firehoses {
			report = append(report, []string{term.Greenf("#%d", index), f.Name})
			index++
		}

		spinner.Stop()
		fmt.Printf("\nShowing %d firehoses", len(firehoses))
		printer.Table(os.Stdout, report)

		return nil
	}

	cmd.Flags().IntVarP(&limit, "limit", "L", 30, "Maximum number of firehoses to fetch (default 30)")

	cmd.Flags().StringVarP(&project, "project", "n", "", "Name of the project")
	cmd.MarkFlagRequired("project")

	return cmd
}
