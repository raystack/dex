package project

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-openapi/strfmt"
	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/odpf/dex/generated/models"
	"github.com/odpf/dex/pkg/errors"
)

const (
	pathParamSlug   = "projectSlug"
	headerProjectID = "X-Shield-Project"
)

func Routes(shield shieldv1beta1.ShieldServiceClient) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", handleListProjects(shield))
		r.Get("/{projectSlug}", handleGetProject(shield))
	}
}

func GetProject(r *http.Request, shieldClient shieldv1beta1.ShieldServiceClient) (*shieldv1beta1.Project, error) {
	projectID := strings.TrimSpace(r.Header.Get(headerProjectID))
	projectSlug := chi.URLParam(r, pathParamSlug)

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

func mapShieldProjectToProject(prj *shieldv1beta1.Project) models.Project {
	return models.Project{
		ID:        prj.Id,
		Name:      prj.Name,
		Slug:      prj.Slug,
		Metadata:  prj.Metadata.AsMap(),
		CreatedAt: strfmt.DateTime(prj.CreatedAt.AsTime()),
		UpdatedAt: strfmt.DateTime(prj.UpdatedAt.AsTime()),
	}
}
