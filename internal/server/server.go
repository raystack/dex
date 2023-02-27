package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/salt/mux"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
	sirenv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/siren/v1beta1"
	"go.uber.org/zap"

	"github.com/odpf/dex/internal/server/reqctx"
	"github.com/odpf/dex/internal/server/utils"
	alertsv1 "github.com/odpf/dex/internal/server/v1/alert"
	firehosev1 "github.com/odpf/dex/internal/server/v1/firehose"
	kubernetesv1 "github.com/odpf/dex/internal/server/v1/kubernetes"
	projectsv1 "github.com/odpf/dex/internal/server/v1/project"
)

// Serve initialises all the HTTP API routes, starts listening for requests at addr, and blocks until
// server exits. Server exits gracefully when context is cancelled.
func Serve(ctx context.Context, addr string,
	nrApp *newrelic.Application, logger *zap.Logger,
	shieldClient shieldv1beta1.ShieldServiceClient,
	entropyClient entropyv1beta1.ResourceServiceClient,
	sirenClient sirenv1beta1.SirenServiceClient,
) error {
	alertSvc := &alertsv1.Service{Siren: sirenClient}

	router := chi.NewRouter()
	curRoute := currentRouteGetter(router)
	router.Use(
		newRelicAPM(nrApp, curRoute),
		requestID(),
		reqctx.WithRequestCtx(),
		withOpenCensus(curRoute),
		requestLogger(logger), // nolint
	)

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, map[string]any{
			"message": "pong",
		})
	})

	router.Route("/api", func(r chi.Router) {
		r.Get("/alertTemplates", alertSvc.HandleListTemplates())

		r.Route("/projects", projectsv1.Routes(shieldClient))
		r.Route("/projects/{projectSlug}/firehoses", firehosev1.Routes(entropyClient, shieldClient, alertSvc))
		r.Route("/projects/{projectSlug}/kubernetes", kubernetesv1.Routes(shieldClient, entropyClient))
	})

	logger.Info("starting server", zap.String("addr", addr))
	return mux.Serve(ctx, addr, mux.WithHTTP(router))
}
