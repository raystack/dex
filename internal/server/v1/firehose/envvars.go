package firehose

import (
	"context"
	"fmt"
	"strings"

	"github.com/goto/dex/generated/models"
	"github.com/goto/dex/odin"
)

func (api *firehoseAPI) buildEnvVars(ctx context.Context, firehose *models.Firehose, userID string, fetchStencilURL bool) error {
	streamURN := buildStreamURN(*firehose.Configs.StreamName, firehose.Project)

	if firehose.Configs.EnvVars[configSourceKafkaBrokers] == "" {
		sourceKafkaBroker, err := odin.GetOdinStream(ctx, api.OdinAddr, streamURN)
		if err != nil {
			return fmt.Errorf("error getting odin stream: %w", err)
		}
		firehose.Configs.EnvVars[configSourceKafkaBrokers] = sourceKafkaBroker
	}

	firehose.Configs.EnvVars["SCHEMA_REGISTRY_STENCIL_ENABLE"] = trueString
	if fetchStencilURL || firehose.Configs.EnvVars[configStencilURL] == "" {
		stencilUrls, err := api.getStencilURLs(
			ctx,
			userID,
			firehose.Configs.EnvVars[configSourceKafkaTopic],
			streamURN,
			firehose.Project,
			firehose.Configs.EnvVars["INPUT_SCHEMA_PROTO_CLASS"],
		)
		if err != nil {
			return fmt.Errorf("error getting stencil url: %w", err)
		}

		firehose.Configs.EnvVars[configStencilURL] = stencilUrls
	}

	sinkType := firehose.Configs.EnvVars[configSinkType]
	firehose.Configs.EnvVars = buildEnvVarsBySink(sinkType, firehose.Configs.EnvVars, firehose.Configs)

	return nil
}

func buildEnvVarsBySink(sinkType string, envVars map[string]string, cfg *models.FirehoseConfig) map[string]string {
	// BQ or GCS sink
	if sinkType == bigquerySinkType || sinkType == blobSinkType {
		envVars["_JAVA_OPTIONS"] = "-Xmx1550m -Xms1550m"
		envVars["DLQ_RETRY_FAIL_AFTER_MAX_ATTEMPT_ENABLE"] = trueString
		envVars["DLQ_RETRY_MAX_ATTEMPTS"] = "5"
		envVars["DLQ_SINK_ENABLE"] = trueString
		envVars["DLQ_WRITER_TYPE"] = "BLOB_STORAGE"
		envVars["ERROR_TYPES_FOR_DLQ"] = "DESERIALIZATION_ERROR,INVALID_MESSAGE_ERROR,UNKNOWN_FIELDS_ERROR,SINK_4XX_ERROR,SINK_5XX_ERROR,SINK_UNKNOWN_ERROR,DEFAULT_ERROR"
		envVars["ERROR_TYPES_FOR_RETRY"] = "SINK_5XX_ERROR,SINK_UNKNOWN_ERROR,DEFAULT_ERROR"
		envVars["RETRY_EXPONENTIAL_BACKOFF_INITIAL_MS"] = "100"
		envVars["RETRY_MAX_ATTEMPTS"] = "10"
		envVars["SOURCE_KAFKA_POLL_TIMEOUT_MS"] = "60000"
	}

	// BQ only
	if sinkType == bigquerySinkType {
		envVars["SOURCE_KAFKA_CONSUMER_MODE"] = "async"
		envVars["SINK_POOL_NUM_THREADS"] = "8"
		envVars["SINK_POOL_QUEUE_POLL_TIMEOUT_MS"] = "100"
		envVars["SINK_BIGQUERY_TABLE_PARTITIONING_ENABLE"] = trueString
		envVars["SINK_BIGQUERY_ROW_INSERT_ID_ENABLE"] = trueString
		envVars["SINK_BIGQUERY_CLIENT_READ_TIMEOUT_MS"] = "-1"
		envVars["SINK_BIGQUERY_CLIENT_CONNECT_TIMEOUT_MS"] = "-1"
		envVars["SINK_BIGQUERY_METADATA_NAMESPACE"] = "__kafka_metadata"
		defaultIfEmpty(envVars, "SINK_BIGQUERY_TABLE_NAME", func() string {
			t := envVars[configSourceKafkaTopic]
			t = strings.ReplaceAll(t, ".", "_")
			t = strings.ReplaceAll(t, "-", "_")
			return t
		})
		defaultIfEmpty(envVars, "SINK_BIGQUERY_DATASET_NAME", func() string {
			return *cfg.StreamName
		})
	}

	return envVars
}

func defaultIfEmpty(vars map[string]string, key string, defaultVal func() string) {
	val, exists := vars[key]
	if !exists || val == "" {
		vars[key] = defaultVal()
	}
}
