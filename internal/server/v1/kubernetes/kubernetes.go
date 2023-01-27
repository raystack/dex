package kubernetes

import (
	"net/http"

	"github.com/gorilla/mux"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
)

const (
	pathParamProjectSlug = "projectSlug"
	kindKubernetes       = "kubernetes"

	// shield header names.
	// Refer https://github.com/odpf/shield
	headerProjectID = "X-Shield-Project"
)

func Routes(r *mux.Router, client entropyv1beta1.ResourceServiceClient, shieldClient shieldv1beta1.ShieldServiceClient) {
	r.Handle("/projects/{projectSlug}/kubernetes", handleListKubernetes(client, shieldClient)).Methods(http.MethodGet)
}
