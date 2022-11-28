package firehose

import (
	"fmt"
	"os"
	"strings"

	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

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
			} else if err := validateFirehoseDef(firehoseDef); err != nil {
				return err
			}

			urn := generateFirehoseURN(args[0], firehoseDef.Name)
			resp, err := client.Operations.GetFirehose(&operations.GetFirehoseParams{ProjectSlug: args[0], FirehoseUrn: urn})
			var notFoundErr *operations.GetFirehoseNotFound
			if err != nil && !errors.As(err, notFoundErr) {
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
						Config: firehoseDef.Configs,
					},
				}
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

				created, createErr := client.Operations.CreateFirehose(params)
				if createErr != nil {
					return createErr
				}
				finalVersion = created.GetPayload()
			}

			fmt.Println(finalVersion)
			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "./config.yaml", "Config file path")
	return cmd
}

func validateFirehoseDef(fd models.Firehose) error {
	if strings.TrimSpace(fd.Name) == "" {
		return errors.New("firehose name must be a non-empty string")
	}
	if strings.TrimSpace(fd.Cluster) == "" {
		return errors.New("")
	}
	return nil
}

func readYAMLFile(filePath string, into interface{}) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return yaml.NewDecoder(f).Decode(into)
}

func generateFirehoseURN(project, name string) string {
	parts := []string{"orn", "entropy", "firehose", project, name}
	return strings.Join(parts, ":")
}
