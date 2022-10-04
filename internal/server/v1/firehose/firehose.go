package firehose

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	pathParamURN = "urn"
	kindFirehose = "firehose"
)

type firehoseDefinition struct {
	URN         string          `json:"urn"`
	Name        string          `json:"name"`
	Team        string          `json:"team"`
	Title       string          `json:"title"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Description string          `json:"description"`
	Cluster     string          `json:"cluster"`
	Configs     firehoseConfigs `json:"configs"`
	State       firehoseState   `json:"state"`
}

type firehoseConfigs struct {
	Image                 string            `json:"image"`
	EnvVars               map[string]string `json:"env_vars"`
	Replicas              int               `json:"replicas"`
	SinkType              string            `json:"sink_type"`
	StopDate              *time.Time        `json:"stop_date"`
	Namespace             string            `json:"namespace"`
	TopicName             string            `json:"topic_name"`
	StreamName            string            `json:"stream_name"`
	ConsumerGroupID       string            `json:"consumer_group_id"`
	BootstrapServers      string            `json:"bootstrap_servers"`
	InputSchemaProtoClass string            `json:"input_schema_proto_class"`
}

type firehoseState struct {
	State        string `json:"state"`
	Status       string `json:"status"`
	DeploymentID string `json:"deployment_id"`
}

type moduleConfig struct {
	State        string                  `json:"state"`
	ChartVersion string                  `json:"chart_version"`
	StopTime     *time.Time              `json:"stop_time"`
	Telegraf     map[string]interface{}  `json:"telegraf"`
	Firehose     moduleConfigFirehoseDef `json:"firehose"`
}

type moduleConfigFirehoseDef struct {
	Replicas           int               `json:"replicas"`
	KafkaBrokerAddress string            `json:"kafka_broker_address"`
	KafkaTopic         string            `json:"kafka_topic"`
	KafkaConsumerID    string            `json:"kafka_consumer_id"`
	EnvVariables       map[string]string `json:"env_variables"`
}

func Routes(r *mux.Router, client entropyv1beta1.ResourceServiceClient) {
	r.Handle("/projects/{projectId}/firehoses", listFirehoses(client)).Methods(http.MethodGet)
	r.Handle("/projects/{projectId}/firehoses", createFirehose(client)).Methods(http.MethodPost)
	r.Handle("/projects/{projectId}/firehoses/{urn}", getFirehose(client)).Methods(http.MethodGet)
	r.Handle("/projects/{projectId}/firehoses/{urn}", updateFirehose(client)).Methods(http.MethodPut)
	r.Handle("/projects/{projectId}/firehoses/{urn}", deleteFirehose(client)).Methods(http.MethodDelete)
}

func mapFirehoseToResource(def firehoseDefinition) (*entropyv1beta1.Resource, error) {
	cfg, err := def.Configs.toConfigStruct()
	if err != nil {
		return nil, err
	}

	spec := &entropyv1beta1.ResourceSpec{
		Configs:      cfg,
		Dependencies: nil, // TODO: source?
	}

	return &entropyv1beta1.Resource{
		Urn:     def.URN,
		Kind:    kindFirehose,
		Name:    def.Name,
		Project: "",  // TODO: source?
		Labels:  nil, // TODO: source?
		Spec:    spec,
	}, nil
}

func mapResourceToFirehose(res *entropyv1beta1.Resource) (*firehoseDefinition, error) {
	if res == nil || res.GetSpec() == nil {
		return nil, errors.New("spec is nil")
	}

	confJSON, err := protojson.Marshal(res.GetSpec().GetConfigs())
	if err != nil {
		return nil, err
	}

	var modConf moduleConfig
	if err := json.Unmarshal(confJSON, &modConf); err != nil {
		return nil, err
	}

	def := firehoseDefinition{
		URN:         res.GetUrn(),
		Name:        res.GetName(),
		Team:        res.GetLabels()["team"],
		Title:       "", // TODO: where is this?
		CreatedAt:   res.GetCreatedAt().AsTime(),
		UpdatedAt:   res.GetUpdatedAt().AsTime(),
		Description: "", // TODO: where is this?
		Cluster:     "", // TODO: where is this?
		Configs: firehoseConfigs{
			Image:                 "odpf/entropy",
			EnvVars:               modConf.Firehose.EnvVariables,
			Replicas:              modConf.Firehose.Replicas,
			SinkType:              modConf.Firehose.EnvVariables["SINK_TYPE"],
			StopDate:              modConf.StopTime,
			Namespace:             "firehose",
			TopicName:             modConf.Firehose.KafkaTopic,
			StreamName:            modConf.Firehose.EnvVariables["STREAM_NAME"],
			ConsumerGroupID:       modConf.Firehose.KafkaConsumerID,
			BootstrapServers:      modConf.Firehose.KafkaBrokerAddress,
			InputSchemaProtoClass: modConf.Firehose.EnvVariables["INPUT_SCHEMA_PROTO_CLASS"],
		},
		State: firehoseState{
			State:        modConf.State,
			Status:       res.GetState().GetStatus().String(),
			DeploymentID: "", // TODO: where is this?
		},
	}

	return &def, nil
}

func (fc firehoseConfigs) toConfigStruct() (*structpb.Value, error) {
	entropyConfig := moduleConfig{
		State:        "", // TODO: where do we get this value?
		ChartVersion: "",
		StopTime:     fc.StopDate,
		Telegraf:     nil,
		Firehose: moduleConfigFirehoseDef{
			Replicas:           fc.Replicas,
			KafkaBrokerAddress: fc.BootstrapServers,
			KafkaTopic:         fc.TopicName,
			KafkaConsumerID:    fc.ConsumerGroupID,
			EnvVariables:       fc.EnvVars,
		},
	}

	jsonB, err := json.Marshal(entropyConfig)
	if err != nil {
		return nil, err
	}

	configStruct := structpb.Value{}
	if err := protojson.Unmarshal(jsonB, &configStruct); err != nil {
		return nil, err
	}

	return &configStruct, nil
}
