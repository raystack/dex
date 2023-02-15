package firehose

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-openapi/strfmt"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/odpf/dex/generated/models"
	"github.com/odpf/dex/internal/server/reqctx"
	"github.com/odpf/dex/internal/server/utils"
	"github.com/odpf/dex/pkg/errors"
)

const kindFirehose = "firehose"

type firehoseUpdates struct {
	Description string                `json:"description"`
	Configs     models.FirehoseConfig `json:"configs"`
}

func (api *firehoseAPI) handleGet(w http.ResponseWriter, r *http.Request) {
	urn := chi.URLParam(r, pathParamURN)

	def, err := api.getFirehose(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, def)
}

func (api *firehoseAPI) handleCreate(w http.ResponseWriter, r *http.Request) {
	var def models.Firehose
	if err := utils.ReadJSON(r, &def); err != nil {
		utils.WriteErr(w, err)
		return
	} else if err := sanitiseAndValidate(&def); err != nil {
		utils.WriteErr(w, err)
		return
	}

	reqCtx := reqctx.From(r.Context())
	def.Metadata = &models.FirehoseMetadata{
		CreatedBy:      strfmt.UUID(reqCtx.UserID),
		CreatedByEmail: strfmt.Email(reqCtx.UserEmail),
		UpdatedBy:      strfmt.UUID(reqCtx.UserID),
		UpdatedByEmail: strfmt.Email(reqCtx.UserEmail),
	}

	prj, err := api.getProject(r)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	res, err := mapFirehoseToResource(def, prj)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	rpcReq := &entropyv1beta1.CreateResourceRequest{Resource: res}
	rpcResp, err := api.Entropy.CreateResource(r.Context(), rpcReq)
	if err != nil {
		outErr := errors.ErrInternal

		st := status.Convert(err)
		if st.Code() == codes.AlreadyExists {
			outErr = errors.ErrConflict.WithCausef(st.Message())
		} else if st.Code() == codes.InvalidArgument {
			outErr = errors.ErrInvalid.WithCausef(st.Message())
		}

		utils.WriteErr(w, outErr)
		return
	}

	createdFirehose, err := mapResourceToFirehose(rpcResp.GetResource(), false)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, createdFirehose)
}

func (api *firehoseAPI) handleDelete(w http.ResponseWriter, r *http.Request) {
	urn := chi.URLParam(r, pathParamURN)

	// Ensure that the URN refers to a valid firehose resource.
	if _, err := api.getFirehose(r.Context(), urn); err != nil {
		utils.WriteErr(w, err)
		return
	}

	rpcReq := &entropyv1beta1.DeleteResourceRequest{Urn: urn}
	_, err := api.Entropy.DeleteResource(r.Context(), rpcReq)
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			utils.WriteErr(w, errFirehoseNotFound)
			return
		}
		utils.WriteErr(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusNoContent, nil)
}

func (api *firehoseAPI) handleList(w http.ResponseWriter, r *http.Request) {
	prj, err := api.getProject(r)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	rpcReq := &entropyv1beta1.ListResourcesRequest{
		Kind:    kindFirehose,
		Project: prj.GetSlug(),
	}

	rpcResp, err := api.Entropy.ListResources(r.Context(), rpcReq)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	var arr []models.Firehose
	for _, res := range rpcResp.GetResources() {
		def, err := mapResourceToFirehose(res, true)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}
		arr = append(arr, *def)
	}

	utils.WriteJSON(w, http.StatusOK,
		utils.ListResponse[models.Firehose]{Items: arr})
}

func (api *firehoseAPI) handleUpdate(w http.ResponseWriter, r *http.Request) {
	urn := chi.URLParam(r, pathParamURN)
	reqCtx := reqctx.From(r.Context())

	var updates firehoseUpdates
	if err := utils.ReadJSON(r, &updates); err != nil {
		utils.WriteErr(w, err)
		return
	}

	existingFirehose, err := api.getFirehose(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	labels := makeLabelsMap(*existingFirehose)
	labels["updated_by"] = reqCtx.UserID
	labels["updated_by_email"] = reqCtx.UserEmail
	if updates.Description != "" {
		labels["description"] = updates.Description
	}

	prj, err := api.getProject(r)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	cfgStruct, err := makeConfigStruct(&updates.Configs, prj)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	rpcReq := &entropyv1beta1.UpdateResourceRequest{
		Urn:    existingFirehose.Urn,
		Labels: labels,
		NewSpec: &entropyv1beta1.ResourceSpec{
			Configs: cfgStruct,
		},
	}

	rpcResp, err := api.Entropy.UpdateResource(r.Context(), rpcReq)
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.InvalidArgument {
			utils.WriteErr(w, errors.ErrInvalid.WithCausef(st.Message()))
		} else if st.Code() == codes.NotFound {
			utils.WriteErr(w, errFirehoseNotFound.WithCausef(st.Message()))
		} else {
			utils.WriteErr(w, err)
		}
		return
	}

	updatedFirehose, err := mapResourceToFirehose(rpcResp.GetResource(), false)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, updatedFirehose)
}

func (api *firehoseAPI) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	urn := chi.URLParam(r, pathParamURN)

	diffs, err := api.getRevisions(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, diffs)
}

func (api *firehoseAPI) getRevisions(ctx context.Context, urn string) ([]models.RevisionDiff, error) {
	rpcReq := &entropyv1beta1.GetResourceRevisionsRequest{Urn: urn}
	rpcResp, err := api.Entropy.GetResourceRevisions(ctx, rpcReq)
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			return nil, errFirehoseNotFound.WithCausef(st.Message())
		}
		return nil, err
	}

	prevSpec := []byte("{}")
	var rh []models.RevisionDiff

	marshaller := protojson.MarshalOptions{
		UseProtoNames: true,
	}

	for _, revision := range rpcResp.GetRevisions() {
		var rd models.RevisionDiff

		currentSpec, err := marshaller.Marshal(revision.GetSpec())
		if err != nil {
			return nil, err
		}

		specDiff, err := jsonDiff(prevSpec, currentSpec)
		if err != nil {
			return nil, err
		}

		rd.Labels = revision.GetLabels()
		rd.Reason = revision.GetReason()
		rd.Diff = json.RawMessage(specDiff)
		rd.UpdatedAt = strfmt.DateTime(revision.GetCreatedAt().AsTime())

		rh = append(rh, rd)
		prevSpec = currentSpec
	}

	return rh, nil
}
