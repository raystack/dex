package cli

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type ClientConfig struct {
	Host string `yaml:"host" cmdx:"host"`
}

type OAuthConfig struct {
	Audience     string        `mapstructure:"audience" yaml:"audience"`
	Endpoint     OAuthEndpoint `mapstructure:"endpoint" yaml:"endpoint"`
	ClientID     string        `mapstructure:"client_id" yaml:"client_id"`
	ClientSecret string        `mapstructure:"client_secret" yaml:"client_secret"`
}

type OAuthEndpoint struct {
	AuthURL  string `mapstructure:"auth_url" yaml:"auth_url"`
	TokenURL string `mapstructure:"token_url" yaml:"token_url"`
}

type AuthConfig struct {
	OAuth        OAuthConfig `mapstructure:"oauth" yaml:"oauth"`
	Expiry       int64       `mapstructure:"expiry" yaml:"expiry"`
	AccessToken  string      `mapstructure:"access_token" yaml:"access_token"`
	RefreshToken string      `mapstructure:"refresh_token" yaml:"refresh_token"`
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
