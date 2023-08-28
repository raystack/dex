package firehose

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	entropyv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/entropy/v1beta1"
	shieldv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/shield/v1beta1"
	"github.com/go-chi/chi/v5"
	"github.com/go-openapi/strfmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/goto/dex/generated/models"
	"github.com/goto/dex/internal/server/reqctx"
	"github.com/goto/dex/internal/server/utils"
	"github.com/goto/dex/internal/server/v1/project"
	"github.com/goto/dex/pkg/errors"
)

const (
	kindFirehose = "firehose"
)

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
	ctx := r.Context()
	reqCtx := reqctx.From(ctx)

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
	def.Configs.StopTime = sanitizeFirehoseStopTime(def.Configs.StopTime)

	groupID := def.Group.String()
	groupSlug, err := api.getGroupSlug(ctx, groupID)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	def.Labels = cloneAndMergeMaps(def.Labels, map[string]string{
		labelTitle:       *def.Title,
		labelGroup:       groupID,
		labelTeam:        groupSlug,
		labelStream:      *def.Configs.StreamName,
		labelDescription: def.Description,
	})

	prj, err := project.GetProject(ctx, def.Project, api.Shield)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	def.Project = prj.GetSlug()

	err = api.buildEnvVars(r.Context(), &def, reqCtx.UserID, true)
	if err != nil {
		utils.WriteErr(w, fmt.Errorf("error building env vars: %w", err))
		return
	}

	res, err := mapFirehoseEntropyResource(def, prj)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	entropyCtx := api.addUserMetadata(ctx, reqCtx.UserEmail)
	rpcReq := &entropyv1beta1.CreateResourceRequest{Resource: res}
	rpcResp, err := api.Entropy.CreateResource(entropyCtx, rpcReq)
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
		configSinkType,
		configSourceKafkaTopic,
		configSourceKafkaConsumerGroup,
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

		if topicName != "" && def.Configs.EnvVars[configSourceKafkaTopic] != topicName {
			continue
		}

		_, include := sinkTypes[def.Configs.EnvVars[configSinkType]]
		if len(sinkTypes) > 0 && !include {
			continue
		}

		// only return selected keys to reduce list response-size.
		returnEnv := map[string]string{}
		for _, key := range includeEnv {
			returnEnv[key] = def.Configs.EnvVars[key]
		}
		def.Configs.EnvVars = returnEnv

		arr = append(arr, def)
	}

	utils.WriteJSON(w, http.StatusOK,
		utils.ListResponse[models.Firehose]{Items: arr})
}

func (api *firehoseAPI) handleUpdate(w http.ResponseWriter, r *http.Request) {
	urn := chi.URLParam(r, pathParamURN)
	ctx := r.Context()
	reqCtx := reqctx.From(ctx)

	var updates struct {
		Group       string                `json:"group"`
		Description string                `json:"description"`
		Configs     models.FirehoseConfig `json:"configs"`
	}
	if err := utils.ReadJSON(r, &updates); err != nil {
		utils.WriteErr(w, err)
		return
	} else if err := updates.Configs.Validate(nil); err != nil {
		utils.WriteErr(w, err)
		return
	} else if updates.Group == "" {
		// TODO: move validation to be same with create
		utils.WriteErr(w, errors.ErrInvalid.WithMsgf("group is required"))
		return
	}

	existingFirehose, err := api.getFirehose(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	} else if updates.Configs.DeploymentID != existingFirehose.Configs.DeploymentID {
		utils.WriteErr(w, errors.ErrInvalid.WithMsgf("deployment_id cannot be updated"))
		return
	} else if *(updates.Configs.KubeCluster) != *(existingFirehose.Configs.KubeCluster) {
		utils.WriteErr(w, errors.ErrInvalid.WithMsgf("kube_cluster cannot be updated"))
		return
	}
	existingFirehose.Group = (*strfmt.UUID)(&updates.Group)
	existingFirehose.Description = updates.Description
	existingFirehose.Configs = &updates.Configs
	existingFirehose.Configs.StopTime = sanitizeFirehoseStopTime(updates.Configs.StopTime)

	groupID := existingFirehose.Group.String()
	groupSlug, err := api.getGroupSlug(ctx, groupID)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}
	labels := cloneAndMergeMaps(existingFirehose.Labels, map[string]string{
		labelGroup: groupID,
		labelTeam:  groupSlug,
	})
	if updates.Description != "" {
		labels[labelDescription] = existingFirehose.Description
	}

	err = api.buildEnvVars(r.Context(), &existingFirehose, reqCtx.UserID, true)
	if err != nil {
		utils.WriteErr(w, fmt.Errorf("error building env vars: %w", err))
		return
	}

	cfgStruct, err := makeConfigStruct(existingFirehose.Configs)
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

	entropyCtx := api.addUserMetadata(ctx, reqCtx.UserEmail)
	rpcResp, err := api.Entropy.UpdateResource(entropyCtx, rpcReq)
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

func (api *firehoseAPI) handlePartialUpdate(w http.ResponseWriter, r *http.Request) {
	urn := chi.URLParam(r, pathParamURN)
	ctx := r.Context()
	reqCtx := reqctx.From(ctx)

	var req struct {
		Group       string                        `json:"group"`
		Description string                        `json:"description"`
		Configs     *models.FirehosePartialConfig `json:"configs"`
	}
	if err := utils.ReadJSON(r, &req); err != nil {
		utils.WriteErr(w, err)
		return
	}

	existing, err := api.getFirehose(r.Context(), urn)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	labels := existing.Labels
	if req.Description != "" {
		labels[labelDescription] = req.Description
	}
	if req.Group != "" {
		groupID := req.Group
		groupSlug, err := api.getGroupSlug(ctx, groupID)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}
		labels[labelGroup] = groupID
		labels[labelTeam] = groupSlug
	}

	if req.Configs.Stopped != nil {
		existing.Configs.Stopped = *req.Configs.Stopped
	}

	if req.Configs.Image != "" {
		existing.Configs.Image = req.Configs.Image
	}

	if req.Configs.StreamName != "" {
		existing.Configs.StreamName = &req.Configs.StreamName
	}

	if req.Configs.Replicas > 0 {
		existing.Configs.Replicas = req.Configs.Replicas
	}

	if req.Configs.StopTime != nil {
		if *req.Configs.StopTime == "" {
			existing.Configs.StopTime = nil
		} else {
			t, err := time.Parse(time.RFC3339, *req.Configs.StopTime)
			if err != nil {
				utils.WriteErr(w, errors.ErrInvalid.
					WithMsgf("stop_time must be valid RFC3339 timestamp").
					WithCausef(err.Error()))
				return
			}
			dt := strfmt.DateTime(t)
			existing.Configs.StopTime = &dt
		}
	}

	existing.Configs.EnvVars = cloneAndMergeMaps(
		existing.Configs.EnvVars,
		req.Configs.EnvVars,
	)

	_, hasTopicUpdate := req.Configs.EnvVars[configSourceKafkaTopic]
	err = api.buildEnvVars(r.Context(), &existing, reqCtx.UserID, hasTopicUpdate)
	if err != nil {
		utils.WriteErr(w, fmt.Errorf("error building env vars: %w", err))
		return
	}

	cfgStruct, err := makeConfigStruct(existing.Configs)
	if err != nil {
		utils.WriteErr(w, err)
		return
	}

	rpcReq := &entropyv1beta1.UpdateResourceRequest{
		Urn:    existing.Urn,
		Labels: labels,
		NewSpec: &entropyv1beta1.ResourceSpec{
			Configs: cfgStruct,
		},
	}

	entropyCtx := api.addUserMetadata(ctx, reqCtx.UserEmail)
	rpcResp, err := api.Entropy.UpdateResource(entropyCtx, rpcReq)
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
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"history": diffs,
	})
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

	revisions := rpcResp.GetRevisions()
	revisionsLen := len(revisions)

	patches := make([]models.RevisionDiff, revisionsLen)
	prevFirehoseJSON := []byte("{}")
	for i := revisionsLen - 1; i >= 0; i-- {
		revision := revisions[i]
		firehose, err := mapEntropySpecAndLabels(models.Firehose{}, revision.GetSpec(), revision.GetLabels())
		if err != nil {
			return nil, err
		}
		currentJSON, err := json.Marshal(firehose)
		if err != nil {
			return nil, err
		}

		var revisionDiff map[string]interface{}
		if revision.Reason != "action:create" {
			revisionDiff, err = jsonDiff(prevFirehoseJSON, currentJSON)
			if err != nil {
				return nil, err
			}
		}

		patch := models.RevisionDiff{
			Reason:    revision.GetReason(),
			UpdatedBy: revision.GetCreatedBy(),
			UpdatedAt: strfmt.DateTime(revision.GetCreatedAt().AsTime()),
			Diff:      revisionDiff,
		}
		patches[i] = patch

		prevFirehoseJSON = currentJSON
	}

	return patches, nil
}

func (api *firehoseAPI) getGroupSlug(ctx context.Context, groupID string) (string, error) {
	resp, err := api.Shield.GetGroup(ctx, &shieldv1beta1.GetGroupRequest{Id: groupID})
	if err != nil {
		return "", fmt.Errorf("error getting group slug: %w", err)
	}

	return resp.Group.GetSlug(), nil
}

func (*firehoseAPI) addUserMetadata(ctx context.Context, userID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx,
		"user-id", userID,
	)
}

func buildStreamURN(streamName, projectSlug string) string {
	return fmt.Sprintf("%s-%s", projectSlug, streamName)
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
