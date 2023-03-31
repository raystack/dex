package firehose

import (
	"context"
	"net/http"
	"time"

	entropyv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/entropy/v1beta1"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/goto/dex/generated/models"
	"github.com/goto/dex/internal/server/reqctx"
	"github.com/goto/dex/internal/server/utils"
	"github.com/goto/dex/pkg/errors"
)

const (
	actionStop        = "stop"
	actionScale       = "scale"
	actionStart       = "start"
	actionUpgrade     = "upgrade"
	actionResetOffset = "reset"
)

func (api *firehoseAPI) handleReset(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		To       string     `json:"to"`
		DateTime *time.Time `json:"date_time"`
	}

	if err := utils.ReadJSON(r, &reqBody); err != nil {
		utils.WriteErr(w, err)
		return
	}

	urn := chi.URLParam(r, pathParamURN)
	updatedFirehose, err := api.executeAction(r.Context(), urn, actionResetOffset, reqBody)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, updatedFirehose)
}

func (api *firehoseAPI) handleScale(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Replicas int `json:"replicas"`
	}

	if err := utils.ReadJSON(r, &reqBody); err != nil {
		utils.WriteErr(w, err)
		return
	}

	urn := chi.URLParam(r, pathParamURN)
	updatedFirehose, err := api.executeAction(r.Context(), urn, actionScale, reqBody)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, updatedFirehose)
}

func (api *firehoseAPI) handleStart(w http.ResponseWriter, r *http.Request) {
	var reqBody struct{}
	if err := utils.ReadJSON(r, &reqBody); err != nil {
		utils.WriteErr(w, err)
		return
	}

	urn := chi.URLParam(r, pathParamURN)
	updatedFirehose, err := api.executeAction(r.Context(), urn, actionStart, reqBody)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, updatedFirehose)
}

func (api *firehoseAPI) handleStop(w http.ResponseWriter, r *http.Request) {
	var reqBody struct{}
	if err := utils.ReadJSON(r, &reqBody); err != nil {
		utils.WriteErr(w, err)
		return
	}

	urn := chi.URLParam(r, pathParamURN)
	updatedFirehose, err := api.executeAction(r.Context(), urn, actionStop, reqBody)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	prj, err := api.getProject(r)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	if err := api.stopAlerts(r.Context(), *updatedFirehose, prj); err != nil {
		utils.WriteErr(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, updatedFirehose)
}

func (api *firehoseAPI) handleUpgrade(w http.ResponseWriter, r *http.Request) {
	var reqBody struct{}
	if err := utils.ReadJSON(r, &reqBody); err != nil {
		utils.WriteErr(w, err)
		return
	}

	urn := chi.URLParam(r, pathParamURN)
	updatedFirehose, err := api.executeAction(r.Context(), urn, actionUpgrade, reqBody)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, updatedFirehose)
}

func (api *firehoseAPI) executeAction(ctx context.Context, urn, actionType string, params any) (*models.Firehose, error) {
	reqCtx := reqctx.From(ctx)

	paramStruct, err := utils.GoValToProtoStruct(params)
	if err != nil {
		return nil, err
	}

	// Ensure that the URN refers to a valid firehose resource.
	existingFirehose, err := api.getFirehose(ctx, urn)
	if err != nil {
		return nil, err
	}

	labels := cloneAndMergeMaps(existingFirehose.Labels, map[string]string{
		labelUpdatedBy: reqCtx.UserEmail,
	})

	rpcReq := &entropyv1beta1.ApplyActionRequest{
		Urn:    urn,
		Action: actionType,
		Params: paramStruct,
		Labels: labels,
	}

	rpcResp, err := api.Entropy.ApplyAction(ctx, rpcReq)
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.InvalidArgument {
			return nil, errors.ErrInvalid.WithCausef(st.Message())
		} else if st.Code() == codes.NotFound {
			return nil, errFirehoseNotFound.WithCausef(st.Message())
		}
		return nil, err
	}

	return mapEntropyResourceToFirehose(rpcResp.GetResource(), nil)
}
