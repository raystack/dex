package firehoses

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
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

	cmd.PersistentFlags().DurationP("timeout", "T", 10*time.Second, "Timeout for the operation")
	return cmd
}

func readYAMLFile(filePath string, into interface{}) error {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	jsonB, err := yaml.YAMLToJSON(b)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonB, into)
}

func generateFirehoseURN(project, name string) string {
	parts := []string{"orn", "entropy", "firehose", project, name}
	return strings.Join(parts, ":")
}
