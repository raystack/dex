package auth

import (
	"fmt"
	"os/signal"
	"syscall"

	"github.com/goto/salt/oidc"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

const (
	scopeEmail = "email"
	redirectTo = "http://localhost:5454"
)

func LoginCommand() *cobra.Command {
	var keyFile string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login using your account (using oauth2 provider) or google service account",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			ac, err := LoadConfig()
			if err != nil {
				return err
			}

			var ts oauth2.TokenSource
			if keyFile == "" {
				ts = oidc.NewTokenSource(ctx, ac.oauth2Config(), ac.OAuth.Audience)
			} else {
				var err error
				ts, err = oidc.NewGoogleServiceAccountTokenSource(ctx, keyFile, ac.OAuth.Audience)
				if err != nil {
					return err
				}
			}

			token, err := ts.Token()
			if err != nil {
				return err
			}

			ac.setToken(token)
			if err := saveConfig(*ac); err != nil {
				return err
			}

			fmt.Println("✅ You are successfully logged in.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&keyFile, "google-key-file", "G", "", "Google service account JSON file")
	return cmd
}
