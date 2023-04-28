package compass

import (
	"context"
	"encoding/json"

	"buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/compass/v1beta1/compassv1beta1grpc"
	compassv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/compass/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/goto/dex/pkg/errors"
)

type assetData struct {
	Attributes struct {
		Schemas []Schema `json:"schemas"`
	} `json:"attributes"`
}

type Schema struct {
	Name        string `json:"name"`
	NamespaceID string `json:"namespaceId"`
	SchemaID    string `json:"schemaId"`
	VersionID   int    `json:"versionId"`
}

func GetTopicSchema(ctx context.Context, cl compassv1beta1grpc.CompassServiceClient,
	userID, projectSlug, stream, topic string, protoNames []string) (*Schema, error) {
	md := metadata.New(map[string]string{"X-Shield-User-Id": userID})

	res, err := cl.GetAllAssets(
		metadata.NewOutgoingContext(ctx, md),
		&compassv1beta1.GetAllAssetsRequest{
			Q:       topic,
			QFields: "name",
			Types:   "topic",
			Data: map[string]string{
				"attributes.project_id": projectSlug,
				"attributes.stream_urn": stream,
			},
		},
		grpc.Header(&metadata.MD{}),
	)
	if err != nil {
		return nil, err
	}

	schemaSet := makeSet(protoNames)
	for _, asset := range res.GetData() {
		val, err := protojson.Marshal(asset.GetData())
		if err != nil {
			return nil, err
		}

		var ad assetData
		if err := json.Unmarshal(val, &ad); err != nil {
			return nil, err
		}

		for _, schema := range ad.Attributes.Schemas {
			if _, found := schemaSet[schema.Name]; found {
				return &schema, nil
			}
		}
	}

	return nil, errors.ErrNotFound.WithMsgf("no schema found")
}

func makeSet(arr []string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, s := range arr {
		set[s] = struct{}{}
	}
	return set
}
