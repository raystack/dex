package firehoses

import (
	"context"
	"log"
	"time"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/cobra"

	"github.com/odpf/dex/cli/auth"
	"github.com/odpf/dex/cli/config"
	"github.com/odpf/dex/generated/client"
)

type swaggerParams interface {
	SetDefaults()
	SetTimeout(d time.Duration)
	SetContext(ctx context.Context)
}

type swaggerTransport struct {
	runtime.ClientTransport

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
	return tr.ClientTransport.Submit(operation)
}

func initClient(cmd *cobra.Command) *client.DexAPI {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configs: %s", err)
	}

	accessToken, err := auth.Token(cmd.Context())
	if err != nil {
		log.Fatalf("failed to load configs: %s", err)
	}

	r := httptransport.New(cfg.Host, "/api", client.DefaultSchemes)
	r.Context = cmd.Context()
	r.Consumers["application/x-ndjson"] = runtime.ByteStreamConsumer()
	r.DefaultAuthentication = httptransport.BearerToken(accessToken)
	r.EnableConnectionReuse()

	customTr := newSwaggerTransport(cmd, r)
	return client.New(customTr, strfmt.Default)
}

func newSwaggerTransport(cmd *cobra.Command, r runtime.ClientTransport) *swaggerTransport {
	d, err := cmd.Flags().GetDuration("timeout")
	if err != nil || d == 0 {
		d = 10 * time.Second
	}

	return &swaggerTransport{
		ClientTransport: r,
		Context:         cmd.Context(),
		Timeout:         d,
	}
}
