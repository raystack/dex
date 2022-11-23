package firehose

import (
	"log"

	"github.com/MakeNowJust/heredoc"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/odpf/salt/cmdx"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/generated/client"
)

type Config struct {
	Host        string `json:"host"`
	AccessToken string `json:"access_token"`
}

type ConfigLoader interface {
	Load(into interface{}, opts ...cmdx.ConfigLoaderOpts) error
}

func Command(cfgLoader ConfigLoader) *cobra.Command {
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
		viewCommand(cfgLoader),
		listCommand(cfgLoader),
		applyCommand(cfgLoader),
		scaleCommand(cfgLoader),
		startCommand(cfgLoader),
		stopCommand(cfgLoader),
		logsCommand(cfgLoader),
		upgradeCommand(cfgLoader),
		resetOffsetCommand(cfgLoader),
	)
	return cmd
}

func initClient(cfgLoader ConfigLoader) *client.DexAPI {
	var cfg Config
	if err := cfgLoader.Load(&cfg); err != nil {
		log.Fatalf("failed to load firehose configs: %s", err)
	}

	r := httptransport.New(cfg.Host, "/api", client.DefaultSchemes)
	r.DefaultAuthentication = httptransport.BearerToken(cfg.AccessToken)
	return client.New(r, strfmt.Default)
}
