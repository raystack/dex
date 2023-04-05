package config

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/goto/salt/cmdx"

	"github.com/goto/dex/pkg/errors"
)

const configAppName = "dex"

var ErrClientConfigNotFound = errors.New(heredoc.Doc(`
	Dex client config not found.
	Run "dex config init" to initialize a new client config or
	Run "dex help environment" for more information.
`))

type ClientConfig struct {
	Host       string `yaml:"host" cmdx:"host"`
	Secure     bool   `yaml:"secure" cmdx:"secure"`
	PathPrefix string `yaml:"path_prefix" cmdx:"path_prefix"`
}

func Load() (*ClientConfig, error) {
	cfg := cmdx.SetConfig(configAppName)

	var cc ClientConfig
	if err := cfg.Load(&cc); err != nil {
		return nil, ErrClientConfigNotFound
	}
	return &cc, nil
}

func initialise() (string, error) {
	cfg := cmdx.SetConfig(configAppName)
	if err := cfg.Init(&ClientConfig{}); err != nil {
		return "", err
	}
	return cfg.File(), nil
}
