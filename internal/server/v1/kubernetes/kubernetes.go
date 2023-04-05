package kubernetes

import (
	"net/http"
	"strings"

	entropyv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/entropy/v1beta1/entropyv1beta1grpc"
	entropyv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/entropy/v1beta1"
	"github.com/go-chi/chi/v5"
	"github.com/go-openapi/strfmt"

	"github.com/goto/dex/generated/models"
	"github.com/goto/dex/internal/server/utils"
	"github.com/goto/dex/pkg/errors"
)

const kindKubernetes = "kubernetes"

func Routes(entropy entropyv1beta1rpc.ResourceServiceClient) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", handleListKubeClusters(entropy))
	}
}

func handleListKubeClusters(entropy entropyv1beta1rpc.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		tagFilter := q["tag"]

		prjSlug := q.Get("project")
		if prjSlug == "" {
			utils.WriteErr(w, errors.ErrInvalid.WithMsgf("project query param must be specified"))
			return
		}

		rpcReq := &entropyv1beta1.ListResourcesRequest{
			Kind:    kindKubernetes,
			Project: prjSlug,
		}

		rpcResp, err := entropy.ListResources(r.Context(), rpcReq)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		var arr []models.Kubernetes
		for _, kube := range rpcResp.GetResources() {
			if matchAllTags(kube.Labels, tagFilter) {
				arr = append(arr, mapResourceToKubernetes(kube))
			}
		}

		utils.WriteJSON(w, http.StatusOK,
			utils.ListResponse[models.Kubernetes]{Items: arr})
	}
}

func matchAllTags(labels map[string]string, tags []string) bool {
	for _, tag := range tags {
		parts := strings.SplitN(tag, ":", 2) // tag is formatted as key:value
		labelKey, wantVal := parts[0], parts[1]
		if labelVal, exists := labels[labelKey]; !exists || labelVal != wantVal {
			// either key does not exist or has a different value than expected.
			return false
		}
	}
	return true
}

func mapResourceToKubernetes(res *entropyv1beta1.Resource) models.Kubernetes {
	return models.Kubernetes{
		Urn:       res.GetUrn(),
		Name:      res.GetName(),
		CreatedAt: strfmt.DateTime(res.GetCreatedAt().AsTime()),
		UpdatedAt: strfmt.DateTime(res.GetUpdatedAt().AsTime()),
	}
}
