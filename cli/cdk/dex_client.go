package cdk

import (
	"context"
	"log"
	"time"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/cobra"

	"github.com/goto/dex/cli/auth"
	"github.com/goto/dex/cli/config"
	"github.com/goto/dex/generated/client"
)

type swaggerParams interface {
	SetDefaults()
	SetTimeout(d time.Duration)
	SetContext(ctx context.Context)
}

type swaggerTransport struct {
	*httptransport.Runtime

	Context    context.Context
	Timeout    time.Duration
	noDeadline bool
}

func (tr *swaggerTransport) Submit(operation *runtime.ClientOperation) (interface{}, error) {
	if params, ok := operation.Params.(swaggerParams); ok {
		if !tr.noDeadline {
			params.SetTimeout(tr.Timeout)
		}
		params.SetContext(tr.Context)
	}

	v, err := tr.Runtime.Submit(operation)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func NewClient(cmd *cobra.Command) *client.DexAPI {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configs: %s", err)
	}

	accessToken, err := auth.Token(cmd.Context())
	if err != nil {
		log.Fatalf("failed to load configs: %s", err)
	}

	scheme := []string{"http"}
	if cfg.Secure {
		scheme = []string{"https"}
	}

	r := httptransport.New(cfg.Host, cfg.PathPrefix, scheme)
	r.Context = cmd.Context()
	r.Consumers["application/x-ndjson"] = runtime.ByteStreamConsumer()
	r.DefaultAuthentication = httptransport.BearerToken(accessToken)
	r.EnableConnectionReuse()

	customTr := newSwaggerTransport(cmd, r)
	return client.New(customTr, strfmt.Default)
}

func newSwaggerTransport(cmd *cobra.Command, r *httptransport.Runtime) *swaggerTransport {
	d, err := cmd.Flags().GetDuration("timeout")
	if err != nil || d == 0 {
		d = 10 * time.Second
	}

	return &swaggerTransport{
		Runtime: r,
		Context: cmd.Context(),
		Timeout: d,
	}
}
