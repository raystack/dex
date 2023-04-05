package project

import (
	"context"

	shieldv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/shield/v1beta1/shieldv1beta1grpc"
	shieldv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/shield/v1beta1"
	"github.com/go-chi/chi/v5"
	"github.com/go-openapi/strfmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/goto/dex/generated/models"
	"github.com/goto/dex/pkg/errors"
)

const pathParamSlug = "projectSlug"

func Routes(shield shieldv1beta1rpc.ShieldServiceClient) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", handleListProjects(shield))
		r.Get("/{projectSlug}", handleGetProject(shield))
	}
}

func GetProject(ctx context.Context, idOrSlug string, shieldClient shieldv1beta1rpc.ShieldServiceClient) (*shieldv1beta1.Project, error) {
	prj, err := shieldClient.GetProject(ctx, &shieldv1beta1.GetProjectRequest{Id: idOrSlug})
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			return nil, errors.ErrNotFound
		}
		return nil, err
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
