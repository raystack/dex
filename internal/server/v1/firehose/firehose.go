package firehose

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
)

func Routes(r *mux.Router, client entropyv1beta1.ResourceServiceClient) {
	r.Handle("/projects/{projectId}/firehoses", listFirehoses(client)).Methods(http.MethodGet)
	r.Handle("/projects/{projectId}/firehoses", createFirehose(client)).Methods(http.MethodPost)
	r.Handle("/projects/{projectId}/firehoses/{urn}", getFirehose(client)).Methods(http.MethodGet)
	r.Handle("/projects/{projectId}/firehoses/{urn}", updateFirehose(client)).Methods(http.MethodPut)
	r.Handle("/projects/{projectId}/firehoses/{urn}", deleteFirehose(client)).Methods(http.MethodDelete)
}

type entropyClient struct {
}

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
