package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/odpf/salt/cmdx"

	"github.com/odpf/dex/cli"
)

const (
	exitOK    = 0
	exitError = 1
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	execute(ctx)
}

func execute(ctx context.Context) {
	root := cli.New()
	cmd, err := root.ExecuteContextC(ctx)

	if err == nil {
		return
	}

	if cmdx.IsCmdErr(err) {
		if !strings.HasSuffix(err.Error(), "\n") {
			fmt.Println()
		}
		fmt.Println(cmd.UsageString())
		os.Exit(exitOK)
	}

	fmt.Println(err)
	os.Exit(exitError)
}
