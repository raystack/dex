package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/odpf/salt/cmdx"
	"github.com/odpf/salt/term"

	"github.com/odpf/dex/cli"
)

const (
	exitUsageErr   = 1
	exitGeneralErr = 2
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	execute(ctx)
}

func execute(ctx context.Context) {
	root := cli.New(ctx)

	cmd, err := root.ExecuteContextC(ctx)
	if err != nil {
		if cmdx.IsCmdErr(err) {
			if !strings.HasSuffix(err.Error(), "\n") {
				fmt.Println()
			}
			fmt.Println(cmd.UsageString())
			os.Exit(exitUsageErr)
		}

		fmt.Println(term.Redf("Error: %v", err))
		os.Exit(exitGeneralErr)
		return
	}
}
