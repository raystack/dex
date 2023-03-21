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
const (
	confKafkaTopic      = "SOURCE_KAFKA_TOPIC"
	confKafkaBrokers    = "SOURCE_KAFKA_BROKERS"
	confProtoClass      = "INPUT_SCHEMA_PROTO_CLASS"
	confStreamName      = "STREAM_NAME"
	confSinkType        = "SINK_TYPE"
	confKafkaConsumerID = "SOURCE_KAFKA_CONSUMER_GROUP_ID"
)

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

func makeConfigStruct(cfg *models.FirehoseConfig) (*structpb.Value, error) {
	// TODO: handled stop date.

	// Refer: https://odpf.github.io/firehose/advance/generic/
	cfg.EnvVars[confSinkType] = strings.ToUpper(string(*cfg.SinkType))
	cfg.EnvVars[confStreamName] = *cfg.StreamName
	cfg.EnvVars[confProtoClass] = *cfg.InputSchemaProtoClass
	cfg.EnvVars[confKafkaBrokers] = *cfg.BootstrapServers
	cfg.EnvVars[confKafkaTopic] = *cfg.TopicName
	cfg.EnvVars[confKafkaConsumerID] = cfg.ConsumerGroupID

	return utils.GoValToProtoStruct(entropyFirehose.Config{
		Replicas:     int(cfg.Replicas),
		DeploymentID: cfg.DeploymentID,
		EnvVariables: cfg.EnvVars,
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

		sinkType := models.FirehoseSinkType(modConf.EnvVariables[confSinkType])
		streamName := modConf.EnvVariables[confStreamName]
		protoClass := modConf.EnvVariables[confProtoClass]
		bootstrapServers := modConf.EnvVariables[confKafkaBrokers]
		topicName := modConf.EnvVariables[confKafkaTopic]

		firehoseDef.Configs = &models.FirehoseConfig{
			BootstrapServers:      &bootstrapServers,
			ConsumerGroupID:       modConf.EnvVariables[confKafkaConsumerID],
			EnvVars:               modConf.EnvVariables,
			InputSchemaProtoClass: &protoClass,
			Replicas:              float64(modConf.Replicas),
			SinkType:              &sinkType,
			StopDate:              strfmt.DateTime(time.Time{}), // TODO: set proper value
			StreamName:            &streamName,
			TopicName:             &topicName,
			DeploymentID:          modConf.DeploymentID,
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
