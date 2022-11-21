package firehose

import (
	"net/http"

	"github.com/gorilla/mux"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
	sirenv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/siren/v1beta1"

	alertsv1 "github.com/odpf/dex/internal/server/v1/alert"
)

const (
	pathParamURN = "urn"
	kindFirehose = "firehose"

	actionStop        = "stop"
	actionScale       = "scale"
	actionStart       = "start"
	actionResetOffset = "reset"

	// shield header names.
	// Refer https://github.com/odpf/shield
	headerProjectID = "X-Shield-Project"
)

func Routes(r *mux.Router, client entropyv1beta1.ResourceServiceClient, shieldClient shieldv1beta1.ShieldServiceClient,
	sirenClient sirenv1beta1.SirenServiceClient, latestFirehoseVersion string,
) {
	alertSvc := &alertsv1.Service{Siren: sirenClient}

	// read APIs
	r.Handle("/projects/{projectSlug}/firehoses", handleListFirehoses(client)).Methods(http.MethodGet)
	r.Handle("/projects/{projectSlug}/firehoses/{urn}", handleGetFirehose(client)).Methods(http.MethodGet)

	// write APIs
	r.Handle("/projects/{projectSlug}/firehoses", handleCreateFirehose(client, shieldClient)).Methods(http.MethodPost)
	r.Handle("/projects/{projectSlug}/firehoses/{urn}", handleUpdateFirehose(client, shieldClient)).Methods(http.MethodPut)
	r.Handle("/projects/{projectSlug}/firehoses/{urn}", handleDeleteFirehose(client)).Methods(http.MethodDelete)

	r.Handle("/projects/{projectSlug}/firehoses/{urn}/reset", handleResetFirehose(client)).Methods(http.MethodPost)
	r.Handle("/projects/{projectSlug}/firehoses/{urn}/scale", handleScaleFirehose(client)).Methods(http.MethodPost)
	r.Handle("/projects/{projectSlug}/firehoses/{urn}/start", handleStartOrStop(client, false)).Methods(http.MethodPost)
	r.Handle("/projects/{projectSlug}/firehoses/{urn}/stop", handleStartOrStop(client, true)).Methods(http.MethodPost)
	r.Handle("/projects/{projectSlug}/firehoses/{urn}/upgrade", handleUpgradeFirehose(client, shieldClient, latestFirehoseVersion)).Methods(http.MethodPost)
	r.Handle("/projects/{projectSlug}/firehoses/{urn}/logs", handleGetFirehoseLogs(client)).Methods(http.MethodGet)

	// alert APIs
	r.Handle("/projects/{projectSlug}/firehoses/{urn}/alertPolicy", handleGetFirehoseAlertPolicies(client, shieldClient, alertSvc)).Methods(http.MethodGet)
	r.Handle("/projects/{projectSlug}/firehoses/{urn}/alertPolicy", handleUpsertFirehoseAlertPolicies(client, shieldClient, alertSvc)).Methods(http.MethodPut)
	r.Handle("/projects/{projectSlug}/firehoses/{urn}/alerts", handleListFirehoseAlerts(client, shieldClient, alertSvc)).Methods(http.MethodGet)
	r.Handle("/alertTemplates", alertsv1.HandleListAlertTemplates(alertSvc, kindFirehose, suppliedAlertVariableNames)).Methods(http.MethodGet)
}
