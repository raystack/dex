package project

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/odpf/dex/pkg/errors"
)

const (
	pathParamSlug   = "slug"
	headerProjectID = "X-Shield-Project"
)

// Routes installs project management APIs to router.
func Routes(r *mux.Router, shieldClient shieldv1beta1.ShieldServiceClient) {
	r.HandleFunc("/projects", handleListProjects(shieldClient)).Methods(http.MethodGet)
	r.HandleFunc("/projects/{slug}", handleGetProject(shieldClient)).Methods(http.MethodGet)
}

func getProject(r *http.Request, shieldClient shieldv1beta1.ShieldServiceClient) (*shieldv1beta1.Project, error) {
	projectID := strings.TrimSpace(r.Header.Get(headerProjectID))
	projectSlug := mux.Vars(r)[pathParamSlug]

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
