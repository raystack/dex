package project

import (
	"net/http"

	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"

	"github.com/odpf/dex/internal/server/utils"
)

type listResponse[T any] struct {
	Items []T `json:"items"`
}

func handleGetProject(client shieldv1beta1.ShieldServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		prj, err := getProject(r, client)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		utils.WriteJSON(w, http.StatusOK, mapShieldProjectToProject(prj))
	}
}

func handleListProjects(client shieldv1beta1.ShieldServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listReq := &shieldv1beta1.ListProjectsRequest{}

		resp, err := client.ListProjects(r.Context(), listReq)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		projects := listResponse[Project]{Items: []Project{}}
		for _, p := range resp.Projects {
			if p == nil {
				continue
			}
			projects.Items = append(projects.Items, mapShieldProjectToProject(p))
		}

		utils.WriteJSON(w, http.StatusOK, projects)
	}
}
