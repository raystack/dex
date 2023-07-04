package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/goto/salt/cmdx"
	"github.com/goto/salt/term"

	"github.com/goto/dex/cli"
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
				log.Println()
			}
			log.Println(cmd.UsageString())
			os.Exit(exitUsageErr)
		}

		log.Println(term.Redf("Error: %v", err))
		os.Exit(exitGeneralErr)
		return
	}
}
