package project

import (
	"net/http"

	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"

	"github.com/odpf/dex/generated/models"
	"github.com/odpf/dex/internal/server/utils"
)

func handleGetProject(shield shieldv1beta1.ShieldServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		prj, err := GetProject(r, shield)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}
		utils.WriteJSON(w, http.StatusOK, mapShieldProjectToProject(prj))
	}
}

func handleListProjects(shield shieldv1beta1.ShieldServiceClient) http.HandlerFunc {
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
