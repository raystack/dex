package server

import (
	"context"
	"fmt"
	"net/http"

	gorillamux "github.com/gorilla/mux"
	"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/salt/mux"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
	sirenv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/siren/v1beta1"
	"go.uber.org/zap"

	"github.com/odpf/dex/internal/server/reqctx"
	firehosesv1 "github.com/odpf/dex/internal/server/v1/firehose"
	kubernetesv1 "github.com/odpf/dex/internal/server/v1/kubernetes"
	projectsv1 "github.com/odpf/dex/internal/server/v1/project"
)

// Serve initialises all the HTTP API routes, starts listening for requests at addr, and blocks until
// server exits. Server exits gracefully when context is cancelled.
func Serve(ctx context.Context, addr string, nrApp *newrelic.Application, logger *zap.Logger,
	shieldClient shieldv1beta1.ShieldServiceClient,
	entropyClient entropyv1beta1.ResourceServiceClient,
	sirenClient sirenv1beta1.SirenServiceClient,
) error {
	httpRouter := gorillamux.NewRouter()
	httpRouter.Use(nrgorilla.Middleware(nrApp))
	httpRouter.Handle("/ping", http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintf(wr, "pong")
	}))

	httpRouter.Use(
		requestID(),
		reqctx.WithRequestCtx(),
		withOpenCensus(),
		requestLogger(logger), // nolint
	)

	// Setup API routes. Refer swagger.yml
	apiRouter := httpRouter.PathPrefix("/api/").Subrouter()
	projectsv1.Routes(apiRouter, shieldClient)
	firehosesv1.Routes(apiRouter, entropyClient, shieldClient, sirenClient)
	kubernetesv1.Routes(apiRouter, entropyClient, shieldClient)

	logger.Info("starting server", zap.String("addr", addr))
	return mux.Serve(ctx, addr, mux.WithHTTP(httpRouter))
}
