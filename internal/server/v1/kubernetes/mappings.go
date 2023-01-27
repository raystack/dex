package kubernetes

import (
	"time"

	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
)

type kubernetesDefinition struct {
	URN       string    `json:"urn"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	labels    map[string]string
}

func mapResourceToKubernetes(res *entropyv1beta1.Resource) *kubernetesDefinition {
	def := kubernetesDefinition{
		URN:       res.GetUrn(),
		Name:      res.GetName(),
		CreatedAt: res.GetCreatedAt().AsTime(),
		UpdatedAt: res.GetUpdatedAt().AsTime(),
		labels:    res.GetLabels(),
	}

	return &def
}

func (kd kubernetesDefinition) checkTag(tag string) bool {
	v, ok := kd.labels[tag]
	return ok && v == "true"
}
