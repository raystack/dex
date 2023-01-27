package firehose

import (
	"encoding/json"
	"time"

	"github.com/mitchellh/mapstructure"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/odpf/dex/internal/server/reqctx"
	"github.com/odpf/dex/pkg/errors"
)

const resourceDepKey = "kube_cluster"

type firehoseDefinition struct {
	URN         string           `json:"urn"`
	Name        string           `json:"name"`
	Group       string           `json:"group"`
	Title       string           `json:"title"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Description string           `json:"description"`
	KubeCluster string           `json:"kube_cluster"`
	Configs     *firehoseConfigs `json:"configs,omitempty"`
	State       *firehoseState   `json:"state,omitempty"`
	Metadata    firehoseMetadata `json:"metadata"`
}

type firehoseMetadata struct {
	CreatedBy      string `json:"created_by"`
	CreatedByEmail string `json:"created_by_email"`
	UpdatedBy      string `json:"updated_by"`
	UpdatedByEmail string `json:"updated_by_email"`
}

type firehoseConfigs struct {
	EnvVars               map[string]string `json:"env_vars"`
	Replicas              int               `json:"replicas"`
	SinkType              string            `json:"sink_type"`
	StopDate              *time.Time        `json:"stop_date"`
	TopicName             string            `json:"topic_name"`
	StreamName            string            `json:"stream_name"`
	ConsumerGroupID       string            `json:"consumer_group_id"`
	BootstrapServers      string            `json:"bootstrap_servers"`
	InputSchemaProtoClass string            `json:"input_schema_proto_class"`
}

type firehoseState struct {
	State        string                 `json:"state"`
	Status       string                 `json:"status"`
	Output       map[string]interface{} `json:"output,omitempty"`
	DeploymentID string                 `json:"deployment_id"`
}

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
	StopTime time.Time               `json:"stop_time"`
	Telegraf map[string]interface{}  `json:"telegraf"`
	Firehose moduleConfigFirehoseDef `json:"firehose"`
}

type moduleConfigFirehoseDef struct {
	Replicas           int               `json:"replicas,omitempty"`
	KafkaBrokerAddress string            `json:"kafka_broker_address,omitempty"`
	KafkaTopic         string            `json:"kafka_topic,omitempty"`
	KafkaConsumerID    string            `json:"kafka_consumer_id,omitempty"`
	EnvVariables       map[string]string `json:"env_variables,omitempty"`
}

type revisionDiff struct {
	Diff      json.RawMessage   `json:"diff"`
	Labels    map[string]string `json:"labels"`
	Reason    string            `json:"reason"`
	UpdatedAt time.Time         `json:"updated_at"`
}

func mapFirehoseToResource(rCtx reqctx.ReqCtx, def firehoseDefinition, prj *shieldv1beta1.Project) (*entropyv1beta1.Resource, error) {
	if def.Configs == nil {
		return nil, errors.ErrInvalid.WithMsgf("configs must be set")
	}

	cfg, err := def.Configs.toConfigStruct(prj)
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	spec := &entropyv1beta1.ResourceSpec{
		Configs: cfg,
		Dependencies: []*entropyv1beta1.ResourceDependency{
			{Key: resourceDepKey, Value: def.KubeCluster},
		},
	}

	labels := def.getLabels()
	labels.setUpdatedBy(rCtx)
	labels.setCreatedBy(rCtx)
	labelsMap, err := labels.toMap()
	if err != nil {
		return nil, err
	}

	return &entropyv1beta1.Resource{
		Urn:     def.URN,
		Kind:    kindFirehose,
		Name:    def.Name,
		Project: prj.GetSlug(),
		Labels:  labelsMap,
		Spec:    spec,
	}, nil
}

func mapResourceToFirehose(res *entropyv1beta1.Resource, onlyMeta bool) (*firehoseDefinition, error) {
	if res == nil || res.GetSpec() == nil {
		return nil, errors.ErrInternal.WithCausef("spec is nil")
	}

	var modConf moduleConfig
	if err := protoStructToGo(res.GetSpec().GetConfigs(), &modConf); err != nil {
		return nil, err
	}

	labelsMap := res.GetLabels()
	labels, err := toFirehoseLabels(labelsMap)
	if err != nil {
		return nil, err
	}

	var kubeCluster string
	for _, dep := range res.GetSpec().GetDependencies() {
		if dep.GetKey() == resourceDepKey {
			kubeCluster = dep.GetValue()
		}
	}

	def := firehoseDefinition{
		URN:         res.GetUrn(),
		Name:        res.GetName(),
		Title:       labels.Title,
		Group:       labels.Group,
		CreatedAt:   res.GetCreatedAt().AsTime(),
		UpdatedAt:   res.GetUpdatedAt().AsTime(),
		Description: labels.Description,
		KubeCluster: kubeCluster,
		Metadata: firehoseMetadata{
			CreatedBy:      labels.CreatedBy,
			CreatedByEmail: labels.CreatedByEmail,
			UpdatedBy:      labels.UpdatedBy,
			UpdatedByEmail: labels.UpdatedByEmail,
		},
	}

	if !onlyMeta {
		def.Configs = &firehoseConfigs{
			EnvVars:               modConf.Firehose.EnvVariables,
			Replicas:              modConf.Firehose.Replicas,
			SinkType:              modConf.Firehose.EnvVariables["SINK_TYPE"],
			StopDate:              &modConf.StopTime,
			TopicName:             modConf.Firehose.KafkaTopic,
			StreamName:            modConf.Firehose.EnvVariables["STREAM_NAME"],
			ConsumerGroupID:       modConf.Firehose.KafkaConsumerID,
			BootstrapServers:      modConf.Firehose.KafkaBrokerAddress,
			InputSchemaProtoClass: modConf.Firehose.EnvVariables["INPUT_SCHEMA_PROTO_CLASS"],
		}
		def.State = &firehoseState{
			State:  modConf.State,
			Status: res.GetState().GetStatus().String(),
			Output: res.GetState().Output.GetStructValue().AsMap(),
		}
	}

	return &def, nil
}

func (fd firehoseDefinition) getLabels() firehoseLabels {
	return firehoseLabels{
		Title:          fd.Title,
		Group:          fd.Group,
		Description:    fd.Description,
		CreatedBy:      fd.Metadata.CreatedBy,
		CreatedByEmail: fd.Metadata.CreatedByEmail,
		UpdatedBy:      fd.Metadata.UpdatedBy,
		UpdatedByEmail: fd.Metadata.UpdatedByEmail,
	}
}

func (fl *firehoseLabels) toMap() (map[string]string, error) {
	result := map[string]string{}
	err := mapstructure.Decode(fl, &result)
	if err != nil {
		return nil, err
	}
	return result, err
}

func toFirehoseLabels(labels map[string]string) (*firehoseLabels, error) {
	result := firehoseLabels{}
	err := mapstructure.Decode(labels, &result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

func (fl *firehoseLabels) setUpdatedBy(ctx reqctx.ReqCtx) {
	fl.UpdatedBy = ctx.UserID
	fl.UpdatedByEmail = ctx.UserEmail
}

func (fl *firehoseLabels) setCreatedBy(ctx reqctx.ReqCtx) {
	fl.CreatedBy = ctx.UserID
	fl.CreatedByEmail = ctx.UserEmail
}

func (fc *firehoseConfigs) toConfigStruct(prj *shieldv1beta1.Project) (*structpb.Value, error) {
	const defaultState = "RUNNING"

	metadata := prj.GetMetadata().AsMap()
	var telegrafConf map[string]interface{}
	telegrafConfString, ok := metadata["telegraf"].(string)
	if ok {
		_ = json.Unmarshal([]byte(telegrafConfString), &telegrafConf)
	}

	if len(telegrafConf) == 0 {
		telegrafConf = map[string]interface{}{"enabled": false}
	}

	return toProtobufStruct(moduleConfig{
		State:    defaultState,
		StopTime: time.Now().Add(10 * time.Hour),
		Telegraf: telegrafConf,
		Firehose: moduleConfigFirehoseDef{
			Replicas:           fc.Replicas,
			KafkaTopic:         fc.TopicName,
			EnvVariables:       fc.EnvVars,
			KafkaConsumerID:    fc.ConsumerGroupID,
			KafkaBrokerAddress: fc.BootstrapServers,
		},
	})
}

func toProtobufStruct(v interface{}) (*structpb.Value, error) {
	jsonB, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	configStruct := structpb.Value{}
	if err := protojson.Unmarshal(jsonB, &configStruct); err != nil {
		return nil, err
	}

	return &configStruct, nil
}

func protoStructToGo(v *structpb.Value, into interface{}) error {
	structJSON, err := protojson.Marshal(v)
	if err != nil {
		return errors.ErrInternal.WithCausef(err.Error())
	}

	if err := json.Unmarshal(structJSON, into); err != nil {
		return errors.ErrInternal.WithCausef(err.Error())
	}
	return nil
}
