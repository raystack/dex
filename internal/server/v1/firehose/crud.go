package firehose

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	entropyv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/entropy/v1beta1"
	"github.com/go-chi/chi/v5"
	"github.com/go-openapi/strfmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/goto/dex/compass"
	"github.com/goto/dex/generated/models"
	"github.com/goto/dex/internal/server/reqctx"
	"github.com/goto/dex/internal/server/utils"
	"github.com/goto/dex/internal/server/v1/project"
	"github.com/goto/dex/odin"
	"github.com/goto/dex/pkg/errors"
)

const (
	kindFirehose  = "firehose"
	confTopicName = "SOURCE_KAFKA_TOPIC"
)

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
	reqCtx := reqctx.From(r.Context())

	var def models.Firehose
	if err := utils.ReadJSON(r, &def); err != nil {
		utils.WriteErr(w, err)
		return
	} else if err := def.Validate(nil); err != nil {
		utils.WriteErr(w, err)
		return
	} else if def.Project == "" {
		utils.WriteErr(w, errors.ErrInvalid.WithMsgf("project must be specified"))
		return
	}

	def.Labels = cloneAndMergeMaps(def.Labels, map[string]string{
		labelTitle:       *def.Title,
		labelGroup:       def.Group.String(),
		labelStream:      *def.Configs.StreamName,
		labelCreatedBy:   reqCtx.UserEmail,
		labelUpdatedBy:   reqCtx.UserEmail,
		labelDescription: def.Description,
	})

	prj, err := project.GetProject(r.Context(), def.Project, api.Shield)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	// resolve stream_name to kafka clusters.
	streamURN := fmt.Sprintf("%s-%s", prj.GetSlug(), *def.Configs.StreamName)
	sourceKafkaBroker, err := odin.GetOdinStream(r.Context(), api.OdinAddr, streamURN)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	def.Configs.EnvVars[confSourceKafkaBrokerAddr] = sourceKafkaBroker

	if def.Configs.EnvVars[confStencilURL] == "" {
		// resolve stencil URL.
		schema, err := compass.GetTopicSchema(
			r.Context(),
			api.Compass,
			reqCtx.UserID,
			prj.GetSlug(),
			streamURN,
			def.Configs.EnvVars[confTopicName],
			strings.Split(def.Configs.EnvVars[confProtoClassName], ","),
		)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}
		def.Configs.EnvVars[confStencilURL] = api.makeStencilURL(*schema)
	}

	res, err := mapFirehoseEntropyResource(def, prj)
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

	createdFirehose, err := mapEntropyResourceToFirehose(rpcResp.GetResource())
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
	q := r.URL.Query()

	prjSlug := q.Get("project")

	if prjSlug == "" {
		utils.WriteErr(w, errors.ErrInvalid.WithMsgf("project query param is required"))
		return
	}

	prj, err := project.GetProject(r.Context(), prjSlug, api.Shield)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	labelFilter := map[string]string{}
	if filterGroup := q.Get("group"); filterGroup != "" {
		labelFilter[labelGroup] = filterGroup
	}

	if streamName := q.Get("stream_name"); streamName != "" {
		labelFilter[labelStream] = streamName
	}

	rpcReq := &entropyv1beta1.ListResourcesRequest{
		Kind:    kindFirehose,
		Project: prj.GetSlug(),
		Labels:  labelFilter,
	}

	rpcResp, err := api.Entropy.ListResources(r.Context(), rpcReq)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	includeEnv := []string{
		confSinkType,
		confTopicName,
		confSourceKafkaConsumerID,
	}

	topicName := q.Get("topic_name")
	kubeCluster := q.Get("kube_cluster")
	sinkTypes := sinkTypeSet(q.Get("sink_type"))
	var arr []models.Firehose
	for _, res := range rpcResp.GetResources() {
		def, err := mapEntropyResourceToFirehose(res)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		if kubeCluster != "" && *def.Configs.KubeCluster != kubeCluster {
			continue
		}

		if topicName != "" && def.Configs.EnvVars[confTopicName] != topicName {
			continue
		}

		_, include := sinkTypes[def.Configs.EnvVars[confSinkType]]
		if len(sinkTypes) > 0 && !include {
			continue
		}

		// only return selected keys to reduce list response-size.
		returnEnv := map[string]string{}
		for _, key := range includeEnv {
			returnEnv[key] = def.Configs.EnvVars[key]
		}
		def.Configs.EnvVars = returnEnv

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

	labels := cloneAndMergeMaps(existingFirehose.Labels, map[string]string{
		labelUpdatedBy: reqCtx.UserEmail,
	})
	if updates.Description != "" {
		labels[labelDescription] = updates.Description
	}

	cfgStruct, err := makeConfigStruct(&updates.Configs)
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

	updatedFirehose, err := mapEntropyResourceToFirehose(rpcResp.GetResource())
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

	revisions := rpcResp.GetRevisions()

	sort.Slice(revisions, func(i, j int) bool {
		return revisions[i].CreatedAt.AsTime().Before(revisions[j].CreatedAt.AsTime())
	})

	for _, revision := range revisions {
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

func sinkTypeSet(sinkTypes string) map[string]struct{} {
	sinkTypes = strings.TrimSpace(sinkTypes)
	if sinkTypes == "" {
		return nil
	}

	res := map[string]struct{}{}
	for _, st := range strings.Split(sinkTypes, ",") {
		res[strings.ToUpper(st)] = struct{}{}
	}
	return res
}
