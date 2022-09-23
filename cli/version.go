package cli

import (
	"github.com/spf13/cobra"

	"github.com/odpf/dex/pkg/version"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return version.Print()
		},
	}
}
