package firehose

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/compass/v1beta1/compassv1beta1grpc"
	entropyv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/entropy/v1beta1/entropyv1beta1grpc"
	shieldv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/shield/v1beta1/shieldv1beta1grpc"
	sirenv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/siren/v1beta1/sirenv1beta1grpc"
	entropyv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/entropy/v1beta1"
	"github.com/go-chi/chi/v5"
	"github.com/wI2L/jsondiff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/goto/dex/compass"
	"github.com/goto/dex/generated/models"
	alertsv1 "github.com/goto/dex/internal/server/v1/alert"
	"github.com/goto/dex/pkg/errors"
)

const pathParamURN = "urn"

var errFirehoseNotFound = errors.ErrNotFound.WithMsgf("no firehose with given URN")

func Routes(entropy entropyv1beta1rpc.ResourceServiceClient,
	shield shieldv1beta1rpc.ShieldServiceClient,
	alertSvc *alertsv1.Service,
	compassClient compassv1beta1grpc.CompassServiceClient,
	odinAddr string,
	stencilHost string,
) func(chi.Router) {
	api := &firehoseAPI{
		Shield:      shield,
		Entropy:     entropy,
		AlertSvc:    alertSvc,
		Compass:     compassClient,
		OdinAddr:    odinAddr,
		StencilAddr: stencilHost,
	}

	return func(r chi.Router) {
		// CRUD operations
		r.Get("/", api.handleList)
		r.Post("/", api.handleCreate)
		r.Get("/{urn}", api.handleGet)
		r.Put("/{urn}", api.handleUpdate)
		r.Patch("/{urn}", api.handlePartialUpdate)
		r.Delete("/{urn}", api.handleDelete)
		r.Get("/{urn}/logs", api.handleStreamLog)
		r.Get("/{urn}/history", api.handleGetHistory)

		// Firehose Actions
		r.Post("/{urn}/reset", api.handleReset)
		r.Post("/{urn}/scale", api.handleScale)
		r.Put("/{urn}/start", api.handleStart)
		r.Put("/{urn}/stop", api.handleStop)
		r.Put("/{urn}/upgrade", api.handleUpgrade)

		// Alert management
		r.Get("/{urn}/alerts", api.handleListAlerts)
		r.Get("/{urn}/alertPolicy", api.handleGetAlertPolicy)
		r.Put("/{urn}/alertPolicy", api.handleUpsertAlertPolicy)
	}
}

type firehoseAPI struct {
	Compass  compassv1beta1grpc.CompassServiceClient
	Entropy  entropyv1beta1rpc.ResourceServiceClient
	Shield   shieldv1beta1rpc.ShieldServiceClient
	Siren    sirenv1beta1rpc.SirenServiceClient
	AlertSvc *alertsv1.Service

	OdinAddr    string
	StencilAddr string
}

func (api *firehoseAPI) getFirehose(ctx context.Context, firehoseURN string) (*models.Firehose, error) {
	resp, err := api.Entropy.GetResource(ctx, &entropyv1beta1.GetResourceRequest{Urn: firehoseURN})
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			return nil, errFirehoseNotFound.WithCausef(st.Message())
		}
		return nil, err
	} else if resp.GetResource().GetKind() != kindFirehose {
		return nil, errFirehoseNotFound
	}

	return mapEntropyResourceToFirehose(resp.GetResource())
}

func (api *firehoseAPI) makeStencilURL(sc compass.Schema) string {
	// Example: https://stencil-host.com/v1beta1/namespaces/{{namespace}}/schemas/{{schema}}
	schemaPath := fmt.Sprintf("/v1beta1/namespaces/%s/schemas/%s", sc.NamespaceID, sc.SchemaID)
	finalURL := strings.TrimSuffix(strings.TrimSpace(api.StencilAddr), "/") + schemaPath
	return finalURL
}

func jsonDiff(left, right []byte) ([]byte, error) {
	patch, err := jsondiff.CompareJSON(left, right)
	if err != nil {
		return nil, err
	}

	return json.Marshal(patch)
}

// Reference: https://github.com/orgs/odpf/discussions/12
func projectSlugFromURN(urn string) string {
	const urnSeparator = ":"
	parts := strings.Split(urn, urnSeparator)
	if len(parts) < 4 {
		return ""
	}

	// The format is: orn:entropy:<kind>:<project>:<name>
	// Project is at index 3.
	return parts[3]
}
