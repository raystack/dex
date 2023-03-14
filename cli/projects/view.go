package projects

import (
	"time"

	"github.com/goto/salt/printer"
	"github.com/spf13/cobra"

	"github.com/goto/dex/cli/cdk"
	"github.com/goto/dex/generated/client/operations"
)

func viewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "view <project-slug>",
		Short:   "View a project",
		Long:    "Display information about a project",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"show", "get"},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			cl := cdk.NewClient(cmd)
			params := operations.GetProjectBySlugParams{
				Slug: args[0],
			}
			params.SetTimeout(10 * time.Second)

			res, err := cl.Operations.GetProjectBySlug(&params)
			if err != nil {
				return err
			}
			project := res.GetPayload()
			spinner.Stop()

			return cdk.Display(cmd, project, cdk.YAMLFormat)
		},
	}

	return cmd
}
