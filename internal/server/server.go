package server

import (
	"context"
	"fmt"
	"net/http"

	gorillamux "github.com/gorilla/mux"
	"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/salt/mux"
	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
	"go.uber.org/zap"

	projectsv1 "github.com/odpf/dex/internal/server/v1/project"
)

// Serve initialises all the HTTP API routes, starts listening for requests at addr, and blocks until
// server exits. Server exits gracefully when context is cancelled.
func Serve(ctx context.Context, addr string, nrApp *newrelic.Application, logger *zap.Logger,
	shieldClient shieldv1beta1.ShieldServiceClient,
) error {
	httpRouter := gorillamux.NewRouter()
	httpRouter.Use(nrgorilla.Middleware(nrApp))
	httpRouter.Handle("/ping", http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintf(wr, "pong")
	}))

	httpRouter.Use(
		requestID(),
		withOpenCensus(),
		requestLogger(logger), // nolint
	)

	// Setup API routes. Refer swagger.yml
	projectsv1.Routes(httpRouter, shieldClient)

	logger.Info("starting server", zap.String("addr", addr))
	return mux.Serve(ctx, addr, mux.WithHTTP(httpRouter))
}
