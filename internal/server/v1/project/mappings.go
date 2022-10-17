package project

import (
	"time"

	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
)

type Project struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

func mapShieldProjectToProject(prj *shieldv1beta1.Project) Project {
	return Project{
		ID:        prj.Id,
		Name:      prj.Name,
		Slug:      prj.Slug,
		CreatedAt: prj.CreatedAt.AsTime(),
		UpdatedAt: prj.UpdatedAt.AsTime(),
		Metadata:  prj.Metadata.AsMap(),
	}
}
