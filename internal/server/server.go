package server

import (
	"context"
	"fmt"
	"net/http"

	gorillamux "github.com/gorilla/mux"
	"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/salt/mux"
	"go.uber.org/zap"
)

// Serve initialises all the HTTP API routes, starts listening for requests at addr, and blocks until
// server exits. Server exits gracefully when context is cancelled.
func Serve(ctx context.Context, addr string, nrApp *newrelic.Application, logger *zap.Logger) error {
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

	logger.Info("starting server", zap.String("addr", addr))
	return mux.Serve(ctx, addr, mux.WithHTTP(httpRouter))
}
