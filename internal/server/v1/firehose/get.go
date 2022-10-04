package firehose

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/odpf/dex/internal/server/utils"
)

const kindFirehose = "firehose"

func getFirehose(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urn := mux.Vars(r)["urn"]

		getResReq := &entropyv1beta1.GetResourceRequest{Urn: urn}
		resp, err := client.GetResource(r.Context(), getResReq)
		if err != nil || resp.GetResource() == nil {
			// TODO: error handling.
			return
		}

		res := resp.GetResource()
		if res == nil || res.GetKind() != kindFirehose {
			// TODO: treat it as not found.
			return
		}

		def, err := mapSpecToFirehose(res)
		if err != nil {
			// TODO: error handling.
			return
		}

		utils.WriteJSON(w, http.StatusOK, def)
	}
}

func mapSpecToFirehose(res *entropyv1beta1.Resource) (*firehoseDefinition, error) {
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

type moduleConfig struct {
	State        string                 `json:"state"`
	ChartVersion string                 `json:"chart_version"`
	StopTime     *time.Time             `json:"stop_time"`
	Telegraf     map[string]interface{} `json:"telegraf"`
	Firehose     struct {
		Replicas           int               `json:"replicas"`
		KafkaBrokerAddress string            `json:"kafka_broker_address"`
		KafkaTopic         string            `json:"kafka_topic"`
		KafkaConsumerID    string            `json:"kafka_consumer_id"`
		EnvVariables       map[string]string `json:"env_variables"`
	} `json:"firehose"`
}
