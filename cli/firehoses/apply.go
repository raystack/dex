package firehoses

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/cli/cdk"
	"github.com/odpf/dex/generated/client/operations"
	"github.com/odpf/dex/generated/models"
	"github.com/odpf/dex/pkg/errors"
)

func applyCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "apply <project> <filepath>",
		Short: "Create/Update a firehose as described in a file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			client := initClient(cmd)

			var firehoseDef models.Firehose
			if err := readYAMLFile(args[1], &firehoseDef); err != nil {
				return err
			}

			urn := generateFirehoseURN(args[0], firehoseDef.Name)
			getParams := &operations.GetFirehoseParams{ProjectSlug: args[0], FirehoseUrn: urn}
			getParams.WithTimeout(10 * time.Second)

			resp, err := client.Operations.GetFirehose(getParams)
			notFoundErr := &operations.GetFirehoseNotFound{}
			if err != nil && !errors.As(err, &notFoundErr) {
				return err
			}

			var finalVersion *models.Firehose
			if resp != nil {
				// Firehose already exists. Treat this as an update.
				existing := resp.GetPayload()
				params := &operations.UpdateFirehoseParams{
					ProjectSlug: args[0],
					FirehoseUrn: existing.Urn,
					Body: operations.UpdateFirehoseBody{
						Configs: firehoseDef.Configs,
					},
				}
				params.WithTimeout(10 * time.Second)

				updated, updateErr := client.Operations.UpdateFirehose(params)
				if updateErr != nil {
					return updateErr
				}
				finalVersion = updated.GetPayload()
			} else {
				// Firehose does not already exist. Treat this as create.
				params := &operations.CreateFirehoseParams{
					Body:        &firehoseDef,
					ProjectSlug: args[0],
				}
				params.WithTimeout(10 * time.Second)

				created, createErr := client.Operations.CreateFirehose(params)
				if createErr != nil {
					return createErr
				}
				finalVersion = created.GetPayload()
			}
			spinner.Stop()

			return cdk.Display(cmd, finalVersion, cdk.YAMLFormat)
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "./config.yaml", "Config file path")
	return cmd
}

func readYAMLFile(filePath string, into interface{}) error {
	b, err := ioutil.ReadFile(filePath)
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
