package firehose

import (
	"fmt"
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

const (
	labelTitle       = "title"
	labelGroup       = "group"
	labelTeam        = "team"
	labelStream      = "stream_name"
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
				{Key: kubeClusterDependencyKey, Value: *def.Configs.KubeCluster},
			},
		},
	}, nil
}

func makeConfigStruct(cfg *models.FirehoseConfig) (*structpb.Value, error) {
	var stopTime *time.Time
	if strings.ToUpper(cfg.EnvVars[configSinkType]) == logSinkType {
		t := time.Now().UTC().Add(logSinkTTL)
		stopTime = &t
	} else if cfg.StopTime != nil {
		t := time.Time(*cfg.StopTime)
		stopTime = &t
	}

	return utils.GoValToProtoStruct(entropyFirehose.Config{
		Stopped:  cfg.Stopped,
		StopTime: stopTime,
		Replicas: int(cfg.Replicas),
		ChartValues: &entropyFirehose.ChartValues{
			ImageTag: cfg.Image,
		},
		DeploymentID: cfg.DeploymentID,
		EnvVariables: cfg.EnvVars,
	})
}

func mapEntropyResourceToFirehose(res *entropyv1beta1.Resource) (models.Firehose, error) {
	if res == nil || res.GetSpec() == nil {
		return models.Firehose{}, errors.ErrInternal.WithCausef("spec is nil")
	}

	firehose := models.Firehose{
		Urn:       res.GetUrn(),
		Name:      res.GetName(),
		Project:   res.Project,
		CreatedBy: res.CreatedBy,
		UpdatedBy: res.UpdatedBy,
		CreatedAt: strfmt.DateTime(res.GetCreatedAt().AsTime()),
		UpdatedAt: strfmt.DateTime(res.GetUpdatedAt().AsTime()),
		State: &models.FirehoseState{
			Status: res.GetState().GetStatus().String(),
			Output: res.GetState().Output.GetStructValue().AsMap(),
		},
	}

	labels := map[string]string{}
	if err := mapstructure.Decode(res.GetLabels(), &labels); err != nil {
		return firehose, fmt.Errorf("error when decoding: %w", err)
	}

	var err error
	firehose, err = mapEntropySpecAndLabels(firehose, res.GetSpec(), labels)
	if err != nil {
		return firehose, errors.ErrInternal.WithCausef(err.Error())
	}

	return firehose, nil
}

func mapEntropySpecAndLabels(firehose models.Firehose, spec *entropyv1beta1.ResourceSpec, labels map[string]string) (models.Firehose, error) {
	title := labels[labelTitle]
	groupID := strfmt.UUID(labels[labelGroup])

	firehose.Title = &title
	firehose.Group = &groupID
	firehose.Labels = labels
	firehose.Description = labels[labelDescription]

	var modConf entropyFirehose.Config
	if err := utils.ProtoStructToGoVal(spec.GetConfigs(), &modConf); err != nil {
		return firehose, err
	}

	var stopTime *strfmt.DateTime
	if modConf.StopTime != nil {
		dt := strfmt.DateTime(*modConf.StopTime)
		stopTime = &dt
	}

	var kubeCluster string
	for _, dep := range spec.GetDependencies() {
		if dep.GetKey() == kubeClusterDependencyKey {
			kubeCluster = dep.GetValue()
		}
	}

	streamName := labels[labelStream]
	if streamName == "" {
		streamName = modConf.EnvVariables[configStreamName]
	}

	firehose.Configs = &models.FirehoseConfig{
		Image:        modConf.ChartValues.ImageTag,
		EnvVars:      modConf.EnvVariables,
		Stopped:      modConf.Stopped,
		StopTime:     stopTime,
		ResetOffset:  modConf.ResetOffset,
		Replicas:     float64(modConf.Replicas),
		StreamName:   &streamName,
		DeploymentID: modConf.DeploymentID,
		KubeCluster:  &kubeCluster,
	}

	return firehose, nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlphaNumPattern.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func cloneAndMergeMaps(m1, m2 map[string]string) map[string]string {
	res := map[string]string{}
	for k, v := range m1 {
		res[k] = v
	}
	for k, v := range m2 {
		res[k] = v
	}
	return res
}
