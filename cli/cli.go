package cli

import (
	"context"

	"github.com/odpf/salt/cmdx"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "dex",
	Long: `What is dex?`,
}

func Execute(ctx context.Context) {
	rootCmd.PersistentFlags().StringP(configFlag, "c", "", "Override config file")
	rootCmd.AddCommand(
		cmdServe(),
		cmdVersion(),
		cmdShowConfigs(),
	)

	cmdx.SetHelp(rootCmd)
	_ = rootCmd.ExecuteContext(ctx)
}
