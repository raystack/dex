package firehoses

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/goto/salt/printer"
	"github.com/spf13/cobra"

	"github.com/goto/dex/cli/cdk"
	"github.com/goto/dex/generated/client/operations"
	"github.com/goto/dex/pkg/errors"
)

func resetOffsetCommand() *cobra.Command {
	var resetTo, datetime string

	cmd := &cobra.Command{
		Use:   "reset-offset <firehoseURN>",
		Short: "Reset firehose consumption offset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client := cdk.NewClient(cmd)

			params := &operations.ResetOffsetParams{FirehoseUrn: args[0]}

			resetTo = strings.TrimSpace(strings.ToUpper(resetTo))
			switch resetTo {
			case "EARLIEST", "LATEST":
				params.Body.To = &resetTo

			case "DATETIME":
				if datetime == "" {
					return errors.New("--datetime flag must be specified when using 'datetime' as reset target")
				}
				dt, err := strfmt.ParseDateTime(datetime)
				if err != nil {
					return errors.Errorf("invalid --datetime value: %v", err)
				}

				params.Body.To = &resetTo
				params.Body.Datetime = dt

			default:
				return errors.Errorf("unknown reset target: %s", resetTo)
			}

			modifiedFirehose, err := client.Operations.ResetOffset(params)
			if err != nil {
				return err
			}

			return cdk.Display(cmd, modifiedFirehose, func(w io.Writer, v interface{}) error {
				_, err := fmt.Fprintln(w, "Reset offset request accepted. Use view command to check status.")
				return err
			})
		},
	}

	cmd.Flags().StringVar(&resetTo, "to", "datetime", "Reset target (earliest, latest, datetime).")
	cmd.Flags().StringVarP(&datetime, "datetime", "D", "", "Target timestamp in ISO8601 or Unix Epoch format.")
	return cmd
}
