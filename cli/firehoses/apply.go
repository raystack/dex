package firehoses

import (
	"fmt"
	"io"
	"log"

	"github.com/goto/salt/printer"
	"github.com/spf13/cobra"

	"github.com/goto/dex/cli/cdk"
	"github.com/goto/dex/generated/client/operations"
	"github.com/goto/dex/generated/models"
	"github.com/goto/dex/pkg/errors"
)

func applyCommand() *cobra.Command {
	var configFile string
	var onlyCreate bool

	cmd := &cobra.Command{
		Use:   "apply <project> <filepath>",
		Short: "Create/Update a firehose as described in a file",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fn := func(cmd *cobra.Command, args []string) error {
				var firehoseDef models.Firehose
				if err := readYAMLFile(args[1], &firehoseDef); err != nil {
					return err
				} else if err := firehoseDef.Validate(nil); err != nil {
					return err
				}

				urn := generateFirehoseURN(args[0], firehoseDef.Name)

				var existing *models.Firehose
				var err error

				var isUpdate bool
				if !onlyCreate {
					notFoundErr := &operations.GetFirehoseNotFound{}
					existing, err = getFirehose(cmd, args[0], urn)
					if err != nil && !errors.As(err, &notFoundErr) {
						return err
					}
					isUpdate = existing != nil
				}

				var finalVersion *models.Firehose
				if isUpdate {
					// Firehose already exists. Treat this as update.
					finalVersion, err = updateFirehose(cmd, args[0], *existing, firehoseDef)
					if err != nil {
						return errors.Errorf("update failed: %s", err)
					}
				} else {
					// Firehose does not already exist. Treat this as create.
					finalVersion, err = createFirehose(cmd, args[0], firehoseDef)
					if err != nil {
						return errors.Errorf("create failed: %s", err)
					}
				}

				return cdk.Display(cmd, finalVersion, func(w io.Writer, v any) error {
					msg := "Create request placed"
					if isUpdate {
						msg = "Update request placed"
					}
					_, err := fmt.Fprintf(w, "%s.\nUse `dex firehose view %s %s` to check status.\n",
						msg, args[0], finalVersion.Urn)
					return err
				})
			}

			err := fn(cmd, args)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}

	cmd.Flags().BoolVar(&onlyCreate, "create", false, "Allow creation only")
	cmd.Flags().StringVarP(&configFile, "config", "c", "./config.yaml", "Config file path")
	return cmd
}

func createFirehose(cmd *cobra.Command, prjSlug string, def models.Firehose) (*models.Firehose, error) {
	spinner := printer.Spin("Creating new firehose")
	defer spinner.Stop()

	// Firehose does not already exist. Treat this as create.
	params := &operations.CreateFirehoseParams{
		Body:        &def,
		ProjectSlug: prjSlug,
	}

	dexAPI := cdk.NewClient(cmd)
	created, createErr := dexAPI.Operations.CreateFirehose(params)
	if createErr != nil {
		return nil, createErr
	}
	return created.GetPayload(), nil
}

func updateFirehose(cmd *cobra.Command, prjSlug string, existing, updated models.Firehose) (*models.Firehose, error) {
	spinner := printer.Spin(fmt.Sprintf("Updating %s", existing.Urn))
	defer spinner.Stop()

	params := &operations.UpdateFirehoseParams{
		ProjectSlug: prjSlug,
		FirehoseUrn: existing.Urn,
		Body: operations.UpdateFirehoseBody{
			Configs:     updated.Configs,
			Description: updated.Description,
		},
	}

	dexAPI := cdk.NewClient(cmd)
	updateResp, updateErr := dexAPI.Operations.UpdateFirehose(params)
	if updateErr != nil {
		return nil, updateErr
	}
	return updateResp.GetPayload(), nil
}
