package projects

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project <command>",
		Aliases: []string{"p"},
		Short:   "Project management commands.",
		Long:    "You can view projects using this command.",
		Example: heredoc.Doc(`
			$ dex project list
			$ dex project show project-x
		`),
		Annotations: map[string]string{
			"group": "core",
		},
	}

	cmd.AddCommand(
		listCommand(),
		viewCommand(),
	)

	return cmd
}
