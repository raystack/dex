package firehose

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	entropyv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/entropy/v1beta1"
	shieldv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/shield/v1beta1"
	"github.com/go-openapi/strfmt"
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

type moduleConfig struct {
	State    string                  `json:"state"`
	StopTime *time.Time              `json:"stop_time,omitempty"`
	Telegraf map[string]interface{}  `json:"telegraf"`
	Firehose moduleConfigFirehoseDef `json:"firehose"`
}

type moduleConfigFirehoseDef struct {
	Replicas           int               `json:"replicas,omitempty"`
	KafkaBrokerAddress string            `json:"kafka_broker_address,omitempty"`
	KafkaTopic         string            `json:"kafka_topic,omitempty"`
	KafkaConsumerID    string            `json:"kafka_consumer_id,omitempty"`
	EnvVariables       map[string]string `json:"env_variables,omitempty"`
	DeploymentID       string            `json:"deployment_id,omitempty"`
}

func sanitiseAndValidate(def *models.Firehose) error {
	if def == nil {
		return errors.ErrInvalid.WithMsgf("definition is nil")
	}

	def.Title = strings.TrimSpace(def.Title)
	def.Name = strings.TrimSpace(def.Name)
	def.Description = strings.TrimSpace(def.Description)
	def.KubeCluster = strings.TrimSpace(def.KubeCluster)

	if def.Title == "" {
		return errors.ErrInvalid.WithMsgf("title must be set")
	}

	if def.Name == "" {
		def.Name = slugify(def.Title)
	}

	if def.KubeCluster == "" {
		return errors.ErrInvalid.WithMsgf("kube_cluster must be set")
	}

	return nil
}

func mapFirehoseToResource(def models.Firehose, prj *shieldv1beta1.Project) (*entropyv1beta1.Resource, error) {
	cfgStruct, err := makeConfigStruct(def.Configs, prj)
	if err != nil {
		return nil, err
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
				{Key: kubeClusterDependencyKey, Value: def.KubeCluster},
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
		"title":            def.Title,
		"group":            def.Group.String(),
		"description":      def.Description,
		"created_by":       meta.CreatedBy.String(),
		"created_by_email": meta.CreatedByEmail.String(),
		"updated_by":       meta.UpdatedBy.String(),
		"updated_by_email": meta.UpdatedByEmail.String(),
	}
}

func makeConfigStruct(cfg *models.FirehoseConfig, prj *shieldv1beta1.Project) (*structpb.Value, error) {
	switch {
	case cfg.BootstrapServers == nil:
		return nil, errors.ErrInvalid.WithMsgf("bootstrap_servers must be set")

	case cfg.TopicName == nil:
		return nil, errors.ErrInvalid.WithMsgf("topic_name must be set")

	case cfg.ConsumerGroupID == nil:
		return nil, errors.ErrInvalid.WithMsgf("consumer_group_id must be set")

	case cfg.SinkType == nil:
		return nil, errors.ErrInvalid.WithMsgf("sink_type must be set")

	case cfg.StreamName == nil:
		return nil, errors.ErrInvalid.WithMsgf("stream_name must be set")

	case cfg.InputSchemaProtoClass == nil:
		return nil, errors.ErrInvalid.WithMsgf("input_schema_proto_class must be set")
	}

	var stopAt *time.Time
	if cfg.StopDate != "" {
		t, err := time.Parse(time.RFC3339, cfg.StopDate)
		if err != nil {
			return nil, errors.ErrInvalid.WithMsgf("stop date must be valid RFC3339 timestamp")
		}
		stopAt = &t
	}

	if cfg.Replicas == nil {
		replicas := float64(1)
		cfg.Replicas = &replicas
	}

	var telegrafConf map[string]any
	prjMetadata := prj.GetMetadata().AsMap()
	if confStr, ok := prjMetadata["telegraf"].(string); ok {
		_ = json.Unmarshal([]byte(confStr), &telegrafConf)

		// disable telegraf by default.
		if len(telegrafConf) == 0 {
			telegrafConf = map[string]interface{}{"enabled": false}
		}
	}

	cfg.EnvVars["SINK_TYPE"] = string(*cfg.SinkType)
	cfg.EnvVars["STREAM_NAME"] = *cfg.StreamName
	cfg.EnvVars["INPUT_SCHEMA_PROTO_CLASS"] = *cfg.InputSchemaProtoClass

	return utils.GoValToProtoStruct(moduleConfig{
		State:    "RUNNING",
		StopTime: stopAt,
		Telegraf: telegrafConf,
		Firehose: moduleConfigFirehoseDef{
			Replicas:           int(*cfg.Replicas),
			KafkaBrokerAddress: *cfg.BootstrapServers,
			KafkaTopic:         *cfg.TopicName,
			KafkaConsumerID:    *cfg.ConsumerGroupID,
			EnvVariables:       cfg.EnvVars,
			DeploymentID:       cfg.DeploymentID,
		},
	})
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

	firehoseDef := models.Firehose{
		Urn:         res.GetUrn(),
		Name:        res.GetName(),
		Title:       labels.Title,
		Group:       strfmt.UUID(labels.Group),
		CreatedAt:   strfmt.DateTime(res.GetCreatedAt().AsTime()),
		UpdatedAt:   strfmt.DateTime(res.GetUpdatedAt().AsTime()),
		Description: labels.Description,
		KubeCluster: kubeCluster,
		Metadata: &models.FirehoseMetadata{
			CreatedBy:      strfmt.UUID(labels.CreatedBy),
			CreatedByEmail: strfmt.Email(labels.CreatedByEmail),
			UpdatedBy:      strfmt.UUID(labels.UpdatedBy),
			UpdatedByEmail: strfmt.Email(labels.UpdatedByEmail),
		},
	}

	if !onlyMeta {
		var modConf moduleConfig
		if err := utils.ProtoStructToGoVal(res.GetSpec().GetConfigs(), &modConf); err != nil {
			return nil, err
		}

		sinkType := models.FirehoseSinkType(modConf.Firehose.EnvVariables["SINK_TYPE"])
		streamName := modConf.Firehose.EnvVariables["STREAM_NAME"]
		protoClass := modConf.Firehose.EnvVariables["INPUT_SCHEMA_PROTO_CLASS"]

		firehoseDef.Configs = &models.FirehoseConfig{
			BootstrapServers:      &modConf.Firehose.KafkaBrokerAddress,
			ConsumerGroupID:       &modConf.Firehose.KafkaConsumerID,
			EnvVars:               modConf.Firehose.EnvVariables,
			InputSchemaProtoClass: &protoClass,
			Replicas:              nil,
			SinkType:              &sinkType,
			StopDate:              timePtrToStr(modConf.StopTime),
			StreamName:            &streamName,
			TopicName:             &modConf.Firehose.KafkaTopic,
			DeploymentID:          modConf.Firehose.DeploymentID,
		}

		firehoseDef.State = &models.FirehoseState{
			State:  modConf.State,
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

func timePtrToStr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.String()
}
