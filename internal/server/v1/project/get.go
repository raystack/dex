package project

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"

	"github.com/odpf/dex/internal/server/utils"
)

type Project struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

func getProject(client shieldv1beta1.ShieldServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rpcReq := &shieldv1beta1.GetProjectRequest{
			Id: mux.Vars(r)["id"],
		}

		rpcResp, err := client.GetProject(r.Context(), rpcReq)
		if err != nil {
			// TODO: handle error
			return
		} else if rpcResp.GetProject() == nil {
			// TODO: project not found?
			return
		}

		shieldProj := rpcResp.GetProject()

		proj := Project{
			ID:        shieldProj.Id,
			Name:      shieldProj.Name,
			Slug:      shieldProj.Slug,
			CreatedAt: shieldProj.CreatedAt.AsTime(),
			UpdatedAt: shieldProj.UpdatedAt.AsTime(),
			Metadata:  shieldProj.Metadata.AsMap(),
		}
		utils.WriteJSON(w, http.StatusOK, proj)
	}
}
