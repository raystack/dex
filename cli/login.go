package cli

import (
	"os/signal"
	"syscall"

	"github.com/odpf/salt/oidc"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

const (
	scopeEmail = "email"
	redirectTo = "http://localhost:5454"
)

func loginCmd(cdk *CDK) *cobra.Command {
	var keyFile string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login using your account (using oauth2 provider) or google service account",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			var ac AuthConfig
			if err := cdk.Auth.Load(&ac); err != nil {
				return err
			}

			oauth2Conf := &oauth2.Config{
				ClientID:     ac.OAuth.ClientID,
				ClientSecret: ac.OAuth.ClientSecret,
				Scopes:       []string{scopeEmail},
				Endpoint:     ac.OAuth.Endpoint,
				RedirectURL:  redirectTo,
			}

			var ts oauth2.TokenSource
			if keyFile == "" {
				ts = oidc.NewTokenSource(ctx, oauth2Conf, ac.OAuth.Audience)
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

			ac.AccessToken = token.AccessToken
			ac.RefreshToken = token.RefreshToken
			ac.Expiry = token.Expiry.Unix()

			return cdk.Auth.Write(ac)
		},
	}

	cmd.Flags().StringVarP(&keyFile, "google-key-file", "G", "", "Google service account JSON file")
	return cmd
}
