package firehose

import (
	"net/http"

	"github.com/gorilla/mux"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
)

const (
	pathParamURN       = "urn"
	pathParamProjectID = "projectId"

	kindFirehose      = "firehose"
	actionResetOffset = "reset"
)

func Routes(r *mux.Router, client entropyv1beta1.ResourceServiceClient, shieldClient shieldv1beta1.ShieldServiceClient) {
	r.Handle("/projects/{projectId}/firehoses", listFirehoses(client)).Methods(http.MethodGet)
	r.Handle("/projects/{projectId}/firehoses", createFirehose(client, shieldClient)).Methods(http.MethodPost)
	r.Handle("/projects/{projectId}/firehoses/{urn}", getFirehose(client)).Methods(http.MethodGet)
	r.Handle("/projects/{projectId}/firehoses/{urn}", updateFirehose(client)).Methods(http.MethodPut)
	r.Handle("/projects/{projectId}/firehoses/{urn}", deleteFirehose(client)).Methods(http.MethodDelete)
	r.Handle("/projects/{projectId}/firehoses/{urn}/reset", resetOffset(client)).Methods(http.MethodPost)
}
