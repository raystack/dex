package firehose

import (
	"regexp"
	"strings"
	"time"

	entropyv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/entropy/v1beta1"
	shieldv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/shield/v1beta1"
	"github.com/go-openapi/strfmt"
	entropyFirehose "github.com/goto/entropy/modules/firehose"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/goto/dex/generated/models"
	"github.com/goto/dex/internal/server/utils"
	"github.com/goto/dex/pkg/errors"
)

const kubeClusterDependencyKey = "kube_cluster"

// Refer https://odpf.github.io/firehose/advance/generic/
const confStreamName = "STREAM_NAME"

const (
	labelTitle       = "title"
	labelGroup       = "group"
	labelCreatedBy   = "created_by"
	labelUpdatedBy   = "updated_by"
	labelDescription = "description"
)

var nonAlphaNumPattern = regexp.MustCompile("[^a-zA-Z0-9]+")

func mapFirehoseEntropyResource(def models.Firehose, prj *shieldv1beta1.Project) (*entropyv1beta1.Resource, error) {
	cfgStruct, err := makeConfigStruct(def.Configs)
	if err != nil {
		return nil, err
	}

	if def.Name == "" {
		def.Name = slugify(*def.Title)
	}

	return &entropyv1beta1.Resource{
		Urn:     def.Urn,
		Kind:    kindFirehose,
		Name:    def.Name,
		Project: prj.GetSlug(),
		Labels:  def.Labels,
		Spec: &entropyv1beta1.ResourceSpec{
			Configs: cfgStruct,
			Dependencies: []*entropyv1beta1.ResourceDependency{
				{Key: kubeClusterDependencyKey, Value: *def.KubeCluster},
			},
		},
	}, nil
}

func makeConfigStruct(cfg *models.FirehoseConfig) (*structpb.Value, error) {
	// TODO: handled stop date.

	// Refer: https://odpf.github.io/firehose/advance/generic/
	cfg.EnvVars[confStreamName] = *cfg.StreamName

	return utils.GoValToProtoStruct(entropyFirehose.Config{
		Replicas:     int(cfg.Replicas),
		DeploymentID: cfg.DeploymentID,
		EnvVariables: cfg.EnvVars,
	})
}

func mapEntropyResourceToFirehose(res *entropyv1beta1.Resource, onlyMeta bool) (*models.Firehose, error) {
	if res == nil || res.GetSpec() == nil {
		return nil, errors.ErrInternal.WithCausef("spec is nil")
	}

	labels := map[string]string{}
	if err := mapstructure.Decode(res.GetLabels(), &labels); err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	var kubeCluster string
	for _, dep := range res.GetSpec().GetDependencies() {
		if dep.GetKey() == kubeClusterDependencyKey {
			kubeCluster = dep.GetValue()
		}
	}

	title := labels[labelTitle]
	groupID := strfmt.UUID(labels[labelGroup])

	firehoseDef := models.Firehose{
		Urn:         res.GetUrn(),
		Name:        res.GetName(),
		Title:       &title,
		Group:       &groupID,
		Labels:      labels,
		CreatedAt:   strfmt.DateTime(res.GetCreatedAt().AsTime()),
		UpdatedAt:   strfmt.DateTime(res.GetUpdatedAt().AsTime()),
		Description: labels[labelDescription],
		KubeCluster: &kubeCluster,
	}

	if !onlyMeta {
		var modConf entropyFirehose.Config
		if err := utils.ProtoStructToGoVal(res.GetSpec().GetConfigs(), &modConf); err != nil {
			return nil, err
		}

		streamName := modConf.EnvVariables[confStreamName]

		firehoseDef.Configs = &models.FirehoseConfig{
			Stopped:  false,                        // TODO: set correct value here.
			StopDate: strfmt.DateTime(time.Time{}), // TODO: set proper value

			Image:        modConf.ChartValues.ImageTag,
			EnvVars:      modConf.EnvVariables,
			Replicas:     float64(modConf.Replicas),
			StreamName:   &streamName,
			DeploymentID: modConf.DeploymentID,
		}

		firehoseDef.State = &models.FirehoseState{
			Status: res.GetState().GetStatus().String(),
			Output: res.GetState().Output.GetStructValue().AsMap(),
		}
	}

	return &firehoseDef, nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlphaNumPattern.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func mergeMaps(m1, m2 map[string]string) map[string]string {
	res := map[string]string{}
	for k, v := range m1 {
		res[k] = v
	}
	for k, v := range m2 {
		res[k] = v
	}
	return res
}
