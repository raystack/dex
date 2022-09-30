package project

import (
	"net/http"

	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"

	"github.com/odpf/dex/internal/server/utils"
)

type listResponse[T any] struct {
	Items []T `json:"items"`
}

func listProjects(client shieldv1beta1.ShieldServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listReq := &shieldv1beta1.ListProjectsRequest{}

		resp, err := client.ListProjects(r.Context(), listReq)
		if err != nil {
			// TODO: handle error
			return
		}

		listResponse := listResponse[Project]{Items: []Project{}}
		for _, p := range resp.Projects {
			if p == nil {
				continue
			}

			listResponse.Items = append(listResponse.Items, Project{
				ID:        p.Id,
				Name:      p.Name,
				Slug:      p.Slug,
				CreatedAt: p.CreatedAt.AsTime(),
				UpdatedAt: p.UpdatedAt.AsTime(),
				Metadata:  p.Metadata.AsMap(),
			})
		}

		utils.WriteJSON(w, http.StatusOK, listResponse)
	}
}
