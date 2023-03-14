package kubernetes

import (
	"net/http"

	entropyv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/entropy/v1beta1/entropyv1beta1grpc"
	shieldv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/shield/v1beta1/shieldv1beta1grpc"
	entropyv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/entropy/v1beta1"
	"github.com/go-chi/chi/v5"
	"github.com/go-openapi/strfmt"

	"github.com/goto/dex/generated/models"
	"github.com/goto/dex/internal/server/utils"
	"github.com/goto/dex/internal/server/v1/project"
)

const kindKubernetes = "kubernetes"

func Routes(shield shieldv1beta1rpc.ShieldServiceClient, entropy entropyv1beta1rpc.ResourceServiceClient) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", handleListKubeClusters(shield, entropy))
	}
}

func handleListKubeClusters(shield shieldv1beta1rpc.ShieldServiceClient, entropy entropyv1beta1rpc.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tag := r.URL.Query().Get("tag")

		prj, err := project.GetProject(r, shield)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		rpcReq := &entropyv1beta1.ListResourcesRequest{
			Kind:    kindKubernetes,
			Project: prj.GetSlug(),
		}

		rpcResp, err := entropy.ListResources(r.Context(), rpcReq)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		var arr []models.Kubernetes
		for _, kube := range rpcResp.GetResources() {
			if matchTag(kube, tag) {
				arr = append(arr, mapResourceToKubernetes(kube))
			}
		}
		utils.WriteJSON(w, http.StatusOK,
			utils.ListResponse[models.Kubernetes]{Items: arr})
	}
}

func mapResourceToKubernetes(res *entropyv1beta1.Resource) models.Kubernetes {
	return models.Kubernetes{
		Urn:       res.GetUrn(),
		Name:      res.GetName(),
		CreatedAt: strfmt.DateTime(res.GetCreatedAt().AsTime()),
		UpdatedAt: strfmt.DateTime(res.GetUpdatedAt().AsTime()),
	}
}

func matchTag(res *entropyv1beta1.Resource, tag string) bool {
	v, ok := res.GetLabels()[tag]
	return ok && v == "true"
}
