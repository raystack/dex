package projects

import (
	"log"

	"github.com/MakeNowJust/heredoc"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/cli/auth"
	"github.com/odpf/dex/cli/config"
	"github.com/odpf/dex/generated/client"
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

func initClient(cmd *cobra.Command) *client.DexAPI {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configs: %s", err)
	}

	accessToken, err := auth.Token(cmd.Context())
	if err != nil {
		log.Fatalf("failed to load configs: %s", err)
	}

	r := httptransport.New(cfg.Host, "/api", client.DefaultSchemes)
	r.Context = cmd.Context()
	r.DefaultAuthentication = httptransport.BearerToken(accessToken)
	return client.New(r, strfmt.Default)
}
