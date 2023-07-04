package server

import (
	"context"
	"net/http"
	"time"

	"buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/compass/v1beta1/compassv1beta1grpc"
	entropyv1beta1 "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/entropy/v1beta1/entropyv1beta1grpc"
	optimusv1beta1 "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/optimus/core/v1beta1/corev1beta1grpc"
	shieldv1beta1 "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/shield/v1beta1/shieldv1beta1grpc"
	sirenv1beta1 "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/siren/v1beta1/sirenv1beta1grpc"
	"github.com/go-chi/chi/v5"
	"github.com/goto/salt/mux"
	"github.com/newrelic/go-agent/v3/newrelic"
	"go.uber.org/zap"

	"github.com/goto/dex/internal/server/reqctx"
	"github.com/goto/dex/internal/server/utils"
	alertsv1 "github.com/goto/dex/internal/server/v1/alert"
	firehosev1 "github.com/goto/dex/internal/server/v1/firehose"
	kubernetesv1 "github.com/goto/dex/internal/server/v1/kubernetes"
	optimusv1 "github.com/goto/dex/internal/server/v1/optimus"
	projectsv1 "github.com/goto/dex/internal/server/v1/project"
)

// Serve initialises all the HTTP API routes, starts listening for requests at addr, and blocks until
// server exits. Server exits gracefully when context is cancelled.
func Serve(ctx context.Context, addr string,
	nrApp *newrelic.Application, logger *zap.Logger,
	shieldClient shieldv1beta1.ShieldServiceClient,
	entropyClient entropyv1beta1.ResourceServiceClient,
	sirenClient sirenv1beta1.SirenServiceClient,
	compassClient compassv1beta1grpc.CompassServiceClient,
	optimusClient optimusv1beta1.JobSpecificationServiceClient,
	odinAddr string,
	stencilAddr string,
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

	router.Route("/dex", func(r chi.Router) {
		r.Get("/alertTemplates", alertSvc.HandleListTemplates())
		r.Route("/optimus", optimusv1.Routes(optimusClient))
		r.Route("/projects", projectsv1.Routes(shieldClient))
		r.Route("/firehoses", firehosev1.Routes(entropyClient, shieldClient, alertSvc, compassClient, odinAddr, stencilAddr))
		r.Route("/kubernetes", kubernetesv1.Routes(entropyClient))
	})

	logger.Info("starting server", zap.String("addr", addr))
	return mux.Serve(ctx,
		mux.WithHTTPTarget(addr, &http.Server{
			Handler:        router,
			ReadTimeout:    120 * time.Second,
			WriteTimeout:   120 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}),
		mux.WithGracePeriod(5*time.Second),
	)
}
