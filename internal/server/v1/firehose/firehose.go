package firehose

import (
	"context"
	"net/http"

	entropyv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/entropy/v1beta1/entropyv1beta1grpc"
	shieldv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/shield/v1beta1/shieldv1beta1grpc"
	sirenv1beta1rpc "buf.build/gen/go/gotocompany/proton/grpc/go/gotocompany/siren/v1beta1/sirenv1beta1grpc"
	entropyv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/entropy/v1beta1"
	shieldv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/shield/v1beta1"
	"github.com/go-chi/chi/v5"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/goto/dex/generated/models"
	alertsv1 "github.com/goto/dex/internal/server/v1/alert"
	"github.com/goto/dex/internal/server/v1/project"
	"github.com/goto/dex/pkg/errors"
)

const pathParamURN = "urn"

var errFirehoseNotFound = errors.ErrNotFound.WithMsgf("no firehose with given URN")

func Routes(entropy entropyv1beta1rpc.ResourceServiceClient,
	shield shieldv1beta1rpc.ShieldServiceClient,
	alertSvc *alertsv1.Service,
) func(chi.Router) {
	api := &firehoseAPI{
		Shield:   shield,
		Entropy:  entropy,
		AlertSvc: alertSvc,
	}

	return func(r chi.Router) {
		// CRUD operations
		r.Get("/", api.handleList)
		r.Post("/", api.handleCreate)
		r.Get("/{urn}", api.handleGet)
		r.Put("/{urn}", api.handleUpdate)
		r.Delete("/{urn}", api.handleDelete)
		r.Get("/{urn}/logs", api.handleStreamLog)
		r.Get("/{urn}/history", api.handleGetHistory)

		// Firehose Actions
		r.Post("/{urn}/reset", api.handleReset)
		r.Post("/{urn}/scale", api.handleScale)
		r.Post("/{urn}/start", api.handleStart)
		r.Post("/{urn}/stop", api.handleStop)
		r.Post("/{urn}/upgrade", api.handleUpgrade)

		// Alert management
		r.Get("/{urn}/alerts", api.handleListAlerts)
		r.Get("/{urn}/alertPolicy", api.handleGetAlertPolicy)
		r.Put("/{urn}/alertPolicy", api.handleUpsertAlertPolicy)
	}
}

type firehoseAPI struct {
	Entropy entropyv1beta1rpc.ResourceServiceClient
	Shield  shieldv1beta1rpc.ShieldServiceClient
	Siren   sirenv1beta1rpc.SirenServiceClient

	AlertSvc *alertsv1.Service
}

func (api *firehoseAPI) getProject(r *http.Request) (*shieldv1beta1.Project, error) {
	return project.GetProject(r, api.Shield)
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

	return mapEntropyResourceToFirehose(resp.GetResource(), false)
}

func jsonDiff(left, right []byte) (string, error) {
	differ := gojsondiff.New()
	compare, err := differ.Compare(left, right)
	if err != nil {
		return "", err
	}

	diffString, err := formatter.NewDeltaFormatter().Format(compare)
	if err != nil {
		return "", err
	}

	return diffString, nil
}
