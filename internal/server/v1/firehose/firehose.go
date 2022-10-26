package firehose

import (
	"net/http"

	"github.com/gorilla/mux"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
)

const (
	pathParamURN      = "urn"
	kindFirehose      = "firehose"
	actionResetOffset = "reset"

	// shield header names.
	// Refer https://github.com/odpf/shield
	headerProjectID = "X-Shield-Project"
)

func Routes(r *mux.Router, client entropyv1beta1.ResourceServiceClient, shieldClient shieldv1beta1.ShieldServiceClient) {
	// read APIs
	r.Handle("/projects/{projectSlug}/firehoses", handleListFirehoses(client)).Methods(http.MethodGet)
	r.Handle("/projects/{projectSlug}/firehoses/{urn}", getFirehose(client)).Methods(http.MethodGet)

	// write APIs
	r.Handle("/projects/{projectSlug}/firehoses", handleCreateFirehose(client, shieldClient)).Methods(http.MethodPost)
	r.Handle("/projects/{projectSlug}/firehoses/{urn}", handleUpdateFirehose(client, shieldClient)).Methods(http.MethodPut)
	r.Handle("/projects/{projectSlug}/firehoses/{urn}", handleDeleteFirehose(client)).Methods(http.MethodDelete)
	r.Handle("/projects/{projectSlug}/firehoses/{urn}/reset", handleResetFirehose(client)).Methods(http.MethodPost)
}
