package firehose

import (
	"encoding/json"
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

var nonAlphaNumPattern = regexp.MustCompile("[^a-zA-Z0-9]+")

type firehoseLabels struct {
	Title          string `mapstructure:"title"`
	Group          string `mapstructure:"group"`
	Description    string `mapstructure:"description"`
	CreatedBy      string `mapstructure:"created_by"`
	CreatedByEmail string `mapstructure:"created_by_email"`
	UpdatedBy      string `mapstructure:"updated_by"`
	UpdatedByEmail string `mapstructure:"updated_by_email"`
}

func mapFirehoseToResource(def models.Firehose, prj *shieldv1beta1.Project) (*entropyv1beta1.Resource, error) {
	cfgStruct, err := makeConfigStruct(def.Configs, prj)
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
		Labels:  makeLabelsMap(def),
		Spec: &entropyv1beta1.ResourceSpec{
			Configs: cfgStruct,
			Dependencies: []*entropyv1beta1.ResourceDependency{
				{Key: kubeClusterDependencyKey, Value: *def.KubeCluster},
			},
		},
	}, nil
}

func makeLabelsMap(def models.Firehose) map[string]string {
	var meta models.FirehoseMetadata
	if def.Metadata != nil {
		meta = *def.Metadata
	}

	return map[string]string{
		"title":            *def.Title,
		"group":            def.Group.String(),
		"description":      def.Description,
		"created_by":       meta.CreatedBy.String(),
		"created_by_email": meta.CreatedByEmail.String(),
		"updated_by":       meta.UpdatedBy.String(),
		"updated_by_email": meta.UpdatedByEmail.String(),
	}
}

func makeConfigStruct(cfg *models.FirehoseConfig, prj *shieldv1beta1.Project) (*structpb.Value, error) {
	stopAt := time.Time(cfg.StopDate)

	var telegrafConf map[string]any
	prjMetadata := prj.GetMetadata().AsMap()
	if confStr, ok := prjMetadata["telegraf"].(string); ok {
		_ = json.Unmarshal([]byte(confStr), &telegrafConf)

		// disable telegraf by default.
		if len(telegrafConf) == 0 {
			telegrafConf = map[string]interface{}{"enabled": false}
		}
	}

	cfg.EnvVars["SINK_TYPE"] = strings.ToUpper(string(*cfg.SinkType))
	cfg.EnvVars["STREAM_NAME"] = *cfg.StreamName
	cfg.EnvVars["INPUT_SCHEMA_PROTO_CLASS"] = *cfg.InputSchemaProtoClass

	var entropyFirehoseConfig entropyFirehose.Config
	entropyFirehoseConfig.State = "RUNNING"
	if !stopAt.IsZero() {
		entropyFirehoseConfig.StopTime = &stopAt
	}
	entropyFirehoseConfig.Telegraf = telegrafConf
	entropyFirehoseConfig.Firehose.Replicas = int(cfg.Replicas)
	entropyFirehoseConfig.Firehose.KafkaBrokerAddress = *cfg.BootstrapServers
	entropyFirehoseConfig.Firehose.KafkaTopic = *cfg.TopicName
	entropyFirehoseConfig.Firehose.KafkaConsumerID = cfg.ConsumerGroupID
	entropyFirehoseConfig.Firehose.EnvVariables = cfg.EnvVars
	entropyFirehoseConfig.Firehose.DeploymentID = cfg.DeploymentID

	return utils.GoValToProtoStruct(entropyFirehoseConfig)
}

func mapResourceToFirehose(res *entropyv1beta1.Resource, onlyMeta bool) (*models.Firehose, error) {
	if res == nil || res.GetSpec() == nil {
		return nil, errors.ErrInternal.WithCausef("spec is nil")
	}

	var labels firehoseLabels
	if err := mapstructure.Decode(res.GetLabels(), &labels); err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	var kubeCluster string
	for _, dep := range res.GetSpec().GetDependencies() {
		if dep.GetKey() == kubeClusterDependencyKey {
			kubeCluster = dep.GetValue()
		}
	}

	groupID := strfmt.UUID(labels.Group)
	firehoseDef := models.Firehose{
		Urn:         res.GetUrn(),
		Name:        res.GetName(),
		Title:       &labels.Title,
		Group:       &groupID,
		CreatedAt:   strfmt.DateTime(res.GetCreatedAt().AsTime()),
		UpdatedAt:   strfmt.DateTime(res.GetUpdatedAt().AsTime()),
		Description: labels.Description,
		KubeCluster: &kubeCluster,
		Metadata: &models.FirehoseMetadata{
			CreatedBy:      strfmt.UUID(labels.CreatedBy),
			CreatedByEmail: strfmt.Email(labels.CreatedByEmail),
			UpdatedBy:      strfmt.UUID(labels.UpdatedBy),
			UpdatedByEmail: strfmt.Email(labels.UpdatedByEmail),
		},
	}

	if !onlyMeta {
		var modConf entropyFirehose.Config
		if err := utils.ProtoStructToGoVal(res.GetSpec().GetConfigs(), &modConf); err != nil {
			return nil, err
		}

		var stopTime time.Time
		if modConf.StopTime != nil {
			stopTime = *modConf.StopTime
		}

		sinkType := models.FirehoseSinkType(modConf.Firehose.EnvVariables["SINK_TYPE"])
		streamName := modConf.Firehose.EnvVariables["STREAM_NAME"]
		protoClass := modConf.Firehose.EnvVariables["INPUT_SCHEMA_PROTO_CLASS"]

		firehoseDef.Configs = &models.FirehoseConfig{
			BootstrapServers:      &modConf.Firehose.KafkaBrokerAddress,
			ConsumerGroupID:       modConf.Firehose.KafkaConsumerID,
			EnvVars:               modConf.Firehose.EnvVariables,
			InputSchemaProtoClass: &protoClass,
			Replicas:              1, // TODO: fix this.
			SinkType:              &sinkType,
			StopDate:              strfmt.DateTime(stopTime),
			StreamName:            &streamName,
			TopicName:             &modConf.Firehose.KafkaTopic,
			DeploymentID:          modConf.Firehose.DeploymentID,
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
