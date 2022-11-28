package firehose

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
		Use:     "firehose <command>",
		Aliases: []string{"s"},
		Short:   "Firehose management commands.",
		Long:    "You can create/manage/view firehoses using this command.",
		Example: heredoc.Doc(`
			$ dex firehose list project-x
			$ dex firehose create -c ./config.yaml
		`),
		Annotations: map[string]string{
			"group": "core",
		},
	}

	cmd.AddCommand(
		viewCommand(),
		listCommand(),
		applyCommand(),
		scaleCommand(),
		startCommand(),
		stopCommand(),
		logsCommand(),
		upgradeCommand(),
		resetOffsetCommand(),
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
