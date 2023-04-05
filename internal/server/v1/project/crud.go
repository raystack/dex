package project

import (
	"net/http"

	shieldv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/shield/v1beta1/shieldv1beta1grpc"
	shieldv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/shield/v1beta1"
	"github.com/go-chi/chi/v5"

	"github.com/goto/dex/generated/models"
	"github.com/goto/dex/internal/server/utils"
)

func handleGetProject(shield shieldv1beta1rpc.ShieldServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idOrSlug := chi.URLParam(r, pathParamSlug)

		prj, err := GetProject(r.Context(), idOrSlug, shield)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}
		utils.WriteJSON(w, http.StatusOK, mapShieldProjectToProject(prj))
	}
}

func handleListProjects(shield shieldv1beta1rpc.ShieldServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listReq := &shieldv1beta1.ListProjectsRequest{}

		resp, err := shield.ListProjects(r.Context(), listReq)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		projects := utils.ListResponse[models.Project]{Items: []models.Project{}}
		for _, p := range resp.Projects {
			if p == nil {
				continue
			}
			projects.Items = append(projects.Items, mapShieldProjectToProject(p))
		}

		utils.WriteJSON(w, http.StatusOK, projects)
	}
}
