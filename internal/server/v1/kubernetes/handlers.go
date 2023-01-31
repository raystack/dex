package kubernetes

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/odpf/dex/internal/server/utils"
	"github.com/odpf/dex/pkg/errors"
)

const (
	queryParamTagKey = "tag"
)

type listResponse[T any] struct {
	Items []T `json:"items"`
}

func handleListKubernetes(client entropyv1beta1.ResourceServiceClient, shieldClient shieldv1beta1.ShieldServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		tag := queryParams.Get(queryParamTagKey)

		prj, err := getProject(r, shieldClient)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		rpcReq := &entropyv1beta1.ListResourcesRequest{
			Kind:    kindKubernetes,
			Project: prj.GetSlug(),
		}

		rpcResp, err := client.ListResources(r.Context(), rpcReq)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		var arr []kubernetesDefinition
		for _, res := range rpcResp.GetResources() {
			kubernetesDef := mapResourceToKubernetes(res)
			if tag == "" || kubernetesDef.checkTag(tag) {
				arr = append(arr, *kubernetesDef)
			}
		}

		resp := listResponse[kubernetesDefinition]{Items: arr}
		utils.WriteJSON(w, http.StatusOK, resp)
	}
}

func getProject(r *http.Request, shieldClient shieldv1beta1.ShieldServiceClient) (*shieldv1beta1.Project, error) {
	projectID := strings.TrimSpace(r.Header.Get(headerProjectID))
	projectSlug := mux.Vars(r)[pathParamProjectSlug]

	if projectID == "" {
		// List everything and search by slug.
		projects, err := shieldClient.ListProjects(r.Context(), &shieldv1beta1.ListProjectsRequest{})
		if err != nil {
			return nil, err
		}
		for _, prj := range projects.GetProjects() {
			if prj.GetSlug() == projectSlug {
				return prj, nil
			}
		}
		return nil, errors.ErrNotFound
	}

	// Project ID is available. Use it to fetch the project directly.
	prj, err := shieldClient.GetProject(r.Context(), &shieldv1beta1.GetProjectRequest{Id: projectID})
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			return nil, errors.ErrNotFound
		}
		return nil, err
	} else if prj.GetProject().Slug != projectSlug {
		return nil, errors.ErrNotFound.WithCausef("projectSlug in URL does not match project of given ID")
	}
	return prj.GetProject(), nil
}
