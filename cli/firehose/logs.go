package firehose

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/generated/client/operations"
	"github.com/odpf/dex/pkg/errors"
)

type logChunk struct {
	Data   []byte            `json:"data,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

func logsCommand(cfgLoader ConfigLoader) *cobra.Command {
	var container, pod string
	var follow, previous, timestamps bool
	var tailCount, since int64

	cmd := &cobra.Command{
		Use:   "logs <project> <firehoseURN>",
		Short: "Stream logs from the given firehose processes",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client := initClient(cmd, cfgLoader)

			params := &operations.GetFirehoseLogsParams{
				FirehoseUrn: args[1],
				ProjectSlug: args[0],
				Follow:      &follow,
				Previous:    &previous,
				Timestamps:  &timestamps,
			}

			if container != "" {
				params.Container = &container
			}

			if pod != "" {
				params.Pod = &pod
			}

			if tailCount > 0 {
				params.TailLines = &tailCount
			}

			if since > 0 {
				params.SinceSeconds = &since
			}

			reader, writer := io.Pipe()
			_, err := client.Operations.GetFirehoseLogs(params, writer)
			if err != nil {
				return err
			}

			onLog := func(chunk logChunk) {
				fmt.Println(chunk)
			}
			if err := streamLogs(cmd.Context(), reader, onLog); err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&pod, "pod", "p", "", "Pod identifier to fetch logs from")
	flags.StringVarP(&container, "container", "c", "", "Container identifier to fetch logs from")
	flags.Int64Var(&since, "since", 0, "Fetch logs since (seconds since unix epoch)")
	flags.Int64VarP(&tailCount, "lines", "n", 0, "Fetch last n lines")
	flags.BoolVarP(&follow, "follow", "f", false, "Stream logs continuously until manual exit")
	flags.BoolVarP(&timestamps, "timestamp", "T", false, "Show timestamps")
	flags.BoolVarP(&previous, "previous", "P", false, "Fetch available logs for previous generation")

	return cmd
}

func streamLogs(ctx context.Context, r io.Reader, onLog func(chunk logChunk)) error {
	sc := bufio.NewScanner(r)

	for sc.Scan() {
		if err := sc.Err(); err != nil {
			return err
		}

		logLine := strings.TrimSpace(sc.Text())
		if logLine != "" {
			var chunk logChunk
			if err := json.Unmarshal([]byte(logLine), &chunk); err != nil {
				return fmt.Errorf("failed to parse log-chunk: %w", err)
			}
			onLog(chunk)
		}

		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
	}

	return nil
}
