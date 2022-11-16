package firehose

import (
	"os"

	"github.com/MakeNowJust/heredoc"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/generated/client"
)

func Command() *cobra.Command {
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

func initClient() *client.DexAPI {
	// TODO: should be read from configurations.
	r := httptransport.New(client.DefaultHost, client.DefaultBasePath, client.DefaultSchemes)
	r.DefaultAuthentication = httptransport.BearerToken(os.Getenv("API_ACCESS_TOKEN"))
	return client.New(r, strfmt.Default)
}
