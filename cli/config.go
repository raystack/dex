package cli

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

type ClientConfig struct {
	Host string `yaml:"host" cmdx:"host"`
}

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	Endpoint     oauth2.Endpoint
	Audience     string
}

type AuthConfig struct {
	OAuth        OAuthConfig
	AccessToken  string
	RefreshToken string
	Expiry       int64
}

func configCmd(cdk *CDK) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage dex CLI configuration",
		Annotations: map[string]string{
			"group": "core",
		},
	}

	cmd.AddCommand(configInitCommand(cdk))
	cmd.AddCommand(configListCommand(cdk))
	return cmd
}

func configInitCommand(cdk *CDK) *cobra.Command {
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
			if err := cdk.Config.Init(&ClientConfig{}); err != nil {
				return err
			}

			fmt.Printf("Config created: %v\n", cdk.Config.File())
			return nil
		},
	}
}

func configListCommand(cdk *CDK) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List client configuration settings",
		Example: heredoc.Doc(`
			$ dex config list
		`),
		Annotations: map[string]string{
			"group": "core",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := cdk.Config.Read()
			if err != nil {
				return ErrClientConfigNotFound
			}

			fmt.Println(data)
			return nil
		},
	}
	return cmd
}
