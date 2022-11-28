package config

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage dex CLI configuration",
		Annotations: map[string]string{
			"group": "core",
		},
	}

	cmd.AddCommand(configInitCommand())
	cmd.AddCommand(configListCommand())
	return cmd
}

func configInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize CLI configuration",
		Example: heredoc.Doc(`
			$ dex config init
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := initialise()
			if err != nil {
				return err
			}
			fmt.Printf("Config created: %v\n", file)
			return nil
		},
	}
}

func configListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show client configuration settings",
		Example: heredoc.Doc(`
			$ dex config show
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cc, err := Load()
			if err != nil {
				return err
			}
			return yaml.NewEncoder(os.Stdout).Encode(cc)
		},
	}
	return cmd
}
