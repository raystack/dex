package auth

import (
	"context"
	"time"

	"github.com/odpf/salt/cmdx"
	"golang.org/x/oauth2"

	"github.com/odpf/dex/pkg/errors"
)

type OAuthProviderConfig struct {
	Audience     string        `mapstructure:"audience" yaml:"audience"`
	Endpoint     OAuthEndpoint `mapstructure:"endpoint" yaml:"endpoint"`
	ClientID     string        `mapstructure:"client_id" yaml:"client_id"`
	ClientSecret string        `mapstructure:"client_secret" yaml:"client_secret"`
}

type OAuthEndpoint struct {
	AuthURL  string `mapstructure:"auth_url" yaml:"auth_url"`
	TokenURL string `mapstructure:"token_url" yaml:"token_url"`
}

type Config struct {
	OAuth        OAuthProviderConfig `mapstructure:"oauth" yaml:"oauth"`
	Expiry       int64               `mapstructure:"expiry" yaml:"expiry"`
	AccessToken  string              `mapstructure:"access_token" yaml:"access_token"`
	RefreshToken string              `mapstructure:"refresh_token" yaml:"refresh_token"`
}

func (ac *Config) setToken(t *oauth2.Token) {
	ac.AccessToken = t.AccessToken
	ac.RefreshToken = t.RefreshToken
	ac.Expiry = t.Expiry.Unix()
}

func (ac *Config) oauth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     ac.OAuth.ClientID,
		ClientSecret: ac.OAuth.ClientSecret,
		Scopes:       []string{scopeEmail},
		Endpoint: oauth2.Endpoint{
			AuthURL:   ac.OAuth.Endpoint.AuthURL,
			TokenURL:  ac.OAuth.Endpoint.TokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: redirectTo,
	}
}

func LoadConfig() (*Config, error) {
	var ac Config

	auth := cmdx.SetConfig("auth")
	if err := auth.Load(&ac); err != nil {
		return nil, err
	}
	return &ac, nil
}

func saveConfig(conf Config) error {
	auth := cmdx.SetConfig("auth")
	return auth.Write(conf)
}

// Token returns a valid access token as per current log-in state. Refresh is performed
// if necessary.
func Token(ctx context.Context) (string, error) {
	ac, err := LoadConfig()
	if err != nil {
		return "", err
	}

	curToken := oauth2.Token{
		Expiry:       time.Unix(ac.Expiry, 0),
		AccessToken:  ac.AccessToken,
		RefreshToken: ac.RefreshToken,
	}

	newToken, err := ac.oauth2Config().TokenSource(ctx, &curToken).Token()
	if err != nil {
		return "", err
	}

	if newToken.RefreshToken != ac.RefreshToken {
		idToken, ok := newToken.Extra("id_token").(string)
		if !ok {
			return "", errors.New("id_token not found in token response")
		}
		newToken.AccessToken = idToken

		ac.setToken(newToken)
		if err := saveConfig(*ac); err != nil {
			return "", err
		}
	}

	return newToken.AccessToken, nil
}
