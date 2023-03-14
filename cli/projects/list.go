package projects

import (
	"fmt"
	"io"
	"time"

	"github.com/goto/salt/printer"
	"github.com/goto/salt/term"
	"github.com/spf13/cobra"

	"github.com/goto/dex/cli/cdk"
	"github.com/goto/dex/generated/client/operations"
)

func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects.",
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client := cdk.NewClient(cmd)

			params := operations.ListProjectsParams{}
			params.SetTimeout(10 * time.Second)

			res, err := client.Operations.ListProjects(&params)
			if err != nil {
				return err
			}
			projects := res.GetPayload().Items
			spinner.Stop()

			return cdk.Display(cmd, projects, func(w io.Writer, v interface{}) error {
				report := [][]string{
					{term.Bold("ID"), term.Bold("SLUG"), term.Bold("NAME")},
				}
				for _, prj := range projects {
					report = append(report, []string{prj.ID, prj.Slug, prj.Name})
				}

				fmt.Printf("Showing %d projects\n", len(projects))
				printer.Table(w, report)
				return nil
			})
		},
	}
	return cmd
}
