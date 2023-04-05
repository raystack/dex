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

func createCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <filepath>",
		Short: "Create a firehose as described in a file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fn := func(cmd *cobra.Command, args []string) error {
				filePath := args[0]

				var firehoseDef models.Firehose
				if err := readYAMLFile(filePath, &firehoseDef); err != nil {
					return err
				} else if err := firehoseDef.Validate(nil); err != nil {
					return err
				}

				finalVersion, err := createFirehose(cmd, firehoseDef)
				if err != nil {
					return errors.Errorf("create failed: %s", err)
				}

				return cdk.Display(cmd, finalVersion, func(w io.Writer, v any) error {
					msg := "Create request placed"
					_, err := fmt.Fprintf(w, "%s.\nUse `dex firehose view %s %s` to check status.\n",
						msg, filePath, finalVersion.Urn)
					return err
				})
			}

			err := fn(cmd, args)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	return cmd
}

func updateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <urn> <filepath>",
		Short: "Update a firehose as described in a file",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fn := func(cmd *cobra.Command, args []string) error {
				urn, filePath := args[0], args[1]

				var firehoseDef models.Firehose
				if err := readYAMLFile(filePath, &firehoseDef); err != nil {
					return err
				} else if err := firehoseDef.Validate(nil); err != nil {
					return err
				}

				finalVersion, err := updateFirehose(cmd, urn, firehoseDef)
				if err != nil {
					return errors.Errorf("update failed: %s", err)
				}

				return cdk.Display(cmd, finalVersion, func(w io.Writer, v any) error {
					msg := "Update request placed"
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
	return cmd
}

func createFirehose(cmd *cobra.Command, def models.Firehose) (*models.Firehose, error) {
	spinner := printer.Spin("Creating new firehose")
	defer spinner.Stop()

	// Firehose does not already exist. Treat this as create.
	params := &operations.CreateFirehoseParams{Body: &def}

	dexAPI := cdk.NewClient(cmd)
	created, createErr := dexAPI.Operations.CreateFirehose(params)
	if createErr != nil {
		return nil, createErr
	}
	return created.GetPayload(), nil
}

func updateFirehose(cmd *cobra.Command, urn string, updated models.Firehose) (*models.Firehose, error) {
	spinner := printer.Spin(fmt.Sprintf("Updating %s", urn))
	defer spinner.Stop()

	params := &operations.UpdateFirehoseParams{
		FirehoseUrn: urn,
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
