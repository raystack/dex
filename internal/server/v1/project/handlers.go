package project

import (
	"net/http"

	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/odpf/dex/internal/server/utils"
	"github.com/odpf/dex/pkg/errors"
)

type listResponse[T any] struct {
	Items []T `json:"items"`
}

func handleGetProject(client shieldv1beta1.ShieldServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rpcReq := &shieldv1beta1.GetProjectRequest{
			Id: r.Header.Get(headerProjectID),
		}

		rpcResp, err := client.GetProject(r.Context(), rpcReq)
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.NotFound {
				utils.WriteErr(w, errors.ErrNotFound.WithCausef(st.Message()))
			} else {
				utils.WriteErr(w, err)
			}
			return
		} else if rpcResp.GetProject() == nil {
			utils.WriteErr(w, errors.ErrNotFound)
			return
		}

		shieldProj := rpcResp.GetProject()
		utils.WriteJSON(w, http.StatusOK, mapShieldProjectToProject(shieldProj))
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

		listResponse := listResponse[Project]{Items: []Project{}}
		for _, p := range resp.Projects {
			if p == nil {
				continue
			}

			listResponse.Items = append(listResponse.Items, mapShieldProjectToProject(p))
		}

		utils.WriteJSON(w, http.StatusOK, listResponse)
	}
}
