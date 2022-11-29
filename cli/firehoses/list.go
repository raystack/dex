package firehoses

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/odpf/salt/printer"
	"github.com/odpf/salt/term"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/cli/cdk"
	"github.com/odpf/dex/generated/client/operations"
)

func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project>",
		Short: "List firehoses in the given project.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client := initClient(cmd)

			params := operations.ListFirehosesParams{
				ProjectSlug: args[0],
			}
			params.SetTimeout(10 * time.Second)
			res, err := client.Operations.ListFirehoses(&params)
			if err != nil {
				return err
			}
			firehoses := res.GetPayload().Items
			spinner.Stop()

			return cdk.Display(cmd, firehoses, func(w io.Writer, v interface{}) error {
				report := [][]string{
					{term.Bold("URN"), term.Bold("NAME"), term.Bold("VERSION")},
				}
				for _, f := range firehoses {
					report = append(report, []string{f.Urn, f.Name, f.Configs.Version})
				}

				fmt.Printf("Showing %d firehoses\n", len(firehoses))
				printer.Table(os.Stdout, report)
				return nil
			})
		},
	}
	return cmd
}
