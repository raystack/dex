package firehose

import (
	"context"
	"fmt"
	"strings"

	"github.com/goto/dex/compass"
)

func (api *firehoseAPI) getStencilURLs(
	ctx context.Context,
	userID string,
	topicStr string,
	streamURN string,
	projectSlug string,
	protoClass string,
) (string, error) {
	topics := strings.Split(topicStr, "|")
	urlMap := map[string]bool{} // using map to prevent duplicate
	urls := []string{}
	for _, topic := range topics {
		// resolve stencil URL.
		schema, err := compass.GetTopicSchema(
			ctx,
			api.Compass,
			userID,
			projectSlug,
			streamURN,
			topic,
			[]string{protoClass},
		)
		if err != nil {
			return "", fmt.Errorf("error getting schema for \"%s\": %w", topic, err)
		}

		fullStencilURL := api.makeStencilURL(schema)
		_, exists := urlMap[fullStencilURL]
		if !exists {
			urlMap[fullStencilURL] = true
			urls = append(urls, fullStencilURL)
		}
	}

	return strings.Join(urls, ","), nil
}

func (api *firehoseAPI) makeStencilURL(sc compass.Schema) string {
	// Example: https://stencil-host.com/v1beta1/namespaces/{{namespace}}/schemas/{{schema}}
	schemaPath := fmt.Sprintf("/v1beta1/namespaces/%s/schemas/%s", sc.NamespaceID, sc.SchemaID)
	finalURL := strings.TrimSuffix(strings.TrimSpace(api.StencilAddr), "/") + schemaPath
	return finalURL
}
