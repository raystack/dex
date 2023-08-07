package firehose

import (
	"context"
	"net/http"
	"time"

	entropyv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/entropy/v1beta1"
	"github.com/go-chi/chi/v5"
	entropyFirehose "github.com/goto/entropy/modules/firehose"
	entropyKafka "github.com/goto/entropy/pkg/kafka"
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
	var reqBody entropyKafka.ResetParams
	if err := utils.ReadJSON(r, &reqBody); err != nil {
		utils.WriteErr(w, err)
		return
	}

	urn := chi.URLParam(r, pathParamURN)
	existingFirehose, err := api.getFirehose(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	updatedFirehose, err := api.executeAction(r.Context(), existingFirehose, actionResetOffset, reqBody)
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

	// Ensure that the URN refers to a valid firehose resource.
	urn := chi.URLParam(r, pathParamURN)
	existingFirehose, err := api.getFirehose(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	params := entropyFirehose.ScaleParams{
		Replicas: reqBody.Replicas,
	}
	updatedFirehose, err := api.executeAction(r.Context(), existingFirehose, actionScale, params)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, updatedFirehose)
}

func (api *firehoseAPI) handleStart(w http.ResponseWriter, r *http.Request) {
	// Ensure that the URN refers to a valid firehose resource.
	urn := chi.URLParam(r, pathParamURN)
	existingFirehose, err := api.getFirehose(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	params := entropyFirehose.StartParams{}
	// for LOG sinkType, updating stop_time
	if existingFirehose.Configs.EnvVars[confSinkType] == logSinkType {
		t := time.Now().UTC().Add(logSinkTTL)
		params.StopTime = &t
	}

	updatedFirehose, err := api.executeAction(r.Context(), existingFirehose, actionStart, params)
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

	// Ensure that the URN refers to a valid firehose resource.
	urn := chi.URLParam(r, pathParamURN)
	existingFirehose, err := api.getFirehose(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	updatedFirehose, err := api.executeAction(r.Context(), existingFirehose, actionStop, reqBody)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	if err := api.stopAlerts(r.Context(), updatedFirehose, projectSlugFromURN(urn)); err != nil {
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

	// Ensure that the URN refers to a valid firehose resource.
	urn := chi.URLParam(r, pathParamURN)
	existingFirehose, err := api.getFirehose(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	updatedFirehose, err := api.executeAction(r.Context(), existingFirehose, actionUpgrade, reqBody)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, updatedFirehose)
}

func (api *firehoseAPI) executeAction(ctx context.Context, existingFirehose models.Firehose, actionType string, params any) (models.Firehose, error) {
	reqCtx := reqctx.From(ctx)

	paramStruct, err := utils.GoValToProtoStruct(params)
	if err != nil {
		return models.Firehose{}, err
	}

	rpcReq := &entropyv1beta1.ApplyActionRequest{
		Urn:    existingFirehose.Urn,
		Action: actionType,
		Params: paramStruct,
		Labels: existingFirehose.Labels,
	}
	entropyCtx := api.addUserMetadata(ctx, reqCtx.UserEmail)
	rpcResp, err := api.Entropy.ApplyAction(entropyCtx, rpcReq)
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.InvalidArgument {
			return models.Firehose{}, errors.ErrInvalid.WithMsgf(st.Message())
		} else if st.Code() == codes.NotFound {
			return models.Firehose{}, errFirehoseNotFound.WithMsgf(st.Message())
		}
		return models.Firehose{}, err
	}

	return mapEntropyResourceToFirehose(rpcResp.GetResource())
}
