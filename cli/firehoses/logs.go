package firehoses

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/goto/dex/cli/cdk"
	"github.com/goto/dex/generated/client/operations"
	"github.com/goto/dex/pkg/errors"
)

type logChunk struct {
	Data   []byte            `json:"data,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

func logsCommand() *cobra.Command {
	var container, pod string
	var follow, previous, timestamps bool
	var tailCount, since int64

	cmd := &cobra.Command{
		Use:   "logs <firehoseURN>",
		Short: "Stream logs from the given firehose processes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if follow {
				_ = cmd.Flags().Set("timeout", "10m")
			}

			dexAPI := cdk.NewClient(cmd)

			params := &operations.GetFirehoseLogsParams{
				FirehoseUrn: args[0],
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

			go func() {
				onLog := func(chunk logChunk) {
					fmt.Print(string(chunk.Data))
				}
				if err := streamLogs(cmd.Context(), reader, onLog); err != nil {
					log.Printf("failed: %v", err)
				}
			}()

			_, err := dexAPI.Operations.GetFirehoseLogs(params, writer)
			return err
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&pod, "pod", "p", "", "Pod identifier to fetch logs from")
	flags.StringVarP(&container, "container", "c", "", "Container identifier to fetch logs from")
	flags.Int64Var(&since, "since", 0, "Fetch logs since (seconds since unix epoch)")
	flags.Int64VarP(&tailCount, "lines", "n", 0, "Fetch last n lines")
	flags.BoolVarP(&follow, "follow", "f", false, "Stream logs continuously until manual exit")
	flags.BoolVarP(&timestamps, "timestamp", "t", false, "Show timestamps")
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
