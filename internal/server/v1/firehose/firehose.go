package firehose

import (
	"context"
	"strings"

	"buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/compass/v1beta1/compassv1beta1grpc"
	entropyv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/entropy/v1beta1/entropyv1beta1grpc"
	shieldv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/shield/v1beta1/shieldv1beta1grpc"
	sirenv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/siren/v1beta1/sirenv1beta1grpc"
	entropyv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/entropy/v1beta1"
	"github.com/StewartJingga/gojsondiff"
	"github.com/StewartJingga/gojsondiff/formatter"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
		r.Put("/{urn}/reset", api.handleReset)
		r.Put("/{urn}/scale", api.handleScale)
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

func (api *firehoseAPI) getFirehose(ctx context.Context, firehoseURN string) (models.Firehose, error) {
	var firehose models.Firehose
	resp, err := api.Entropy.GetResource(ctx, &entropyv1beta1.GetResourceRequest{Urn: firehoseURN})
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			return firehose, errFirehoseNotFound.WithCausef(st.Message())
		}
		return firehose, err
	} else if resp.GetResource().GetKind() != kindFirehose {
		return firehose, errFirehoseNotFound
	}

	return mapEntropyResourceToFirehose(resp.GetResource())
}

func jsonDiff(prev, current []byte) (map[string]interface{}, error) {
	differ := &gojsondiff.Differ{
		TextDiffMinimumLength: 1000,
	}
	diff, err := differ.Compare(prev, current)
	if err != nil {
		return nil, err
	}

	diffMap, err := formatter.NewDeltaFormatter().FormatAsJson(diff)
	if err != nil {
		return nil, err
	}

	return diffMap, nil
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
