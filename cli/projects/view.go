package projects

import (
	"time"

	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/cli/cdk"
	"github.com/odpf/dex/generated/client/operations"
)

func viewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "view <id> <project-slug>",
		Short:   "View a project",
		Long:    "Display information about a project",
		Args:    cobra.ExactArgs(2),
		Aliases: []string{"show"},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client := initClient(cmd)

			params := operations.GetProjectBySlugParams{
				Slug:           args[1],
				XShieldProject: args[0],
			}
			params.SetTimeout(10 * time.Second)

			res, err := client.Operations.GetProjectBySlug(&params)
			if err != nil {
				return err
			}
			project := res.GetPayload()
			spinner.Stop()

			return cdk.Display(cmd, project, nil)
		},
	}

	return cmd
}
