package cli

import (
	"errors"

	"github.com/MakeNowJust/heredoc"
)

var (
	ErrClientConfigNotFound = errors.New(heredoc.Doc(`
		Dex client config not found.
		Run "dex config init" to initialize a new client config or
		Run "dex help environment" for more information.
	`))
	ErrClientConfigHostNotFound = errors.New(heredoc.Doc(`
		Dex client config "host" not found.
		Pass dex server host with "--host" flag or 
		set host in dex config.
		Run "dex config <subcommand>" or
		"dex help environment" for more information.
	`))
	ErrClientNotAuthorized = errors.New(heredoc.Doc(`
		Dex auth error. Dex requires an auth header.
		
		Run "dex help auth" for more information.
	`))
)
