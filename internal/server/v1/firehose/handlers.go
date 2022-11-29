package firehose

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/odpf/dex/internal/server/reqctx"
	"github.com/odpf/dex/internal/server/utils"
	alertsv1 "github.com/odpf/dex/internal/server/v1/alert"
	"github.com/odpf/dex/pkg/errors"
)

const (
	firehoseNotFound             = "no firehose with given URN"
	firehoseOutputReleaseNameKey = "release_name"
)

var (
	firehoseLogFilterKeys      = []string{"pod", "container", "since_seconds", "tail_lines", "follow", "previous", "timestamps"}
	suppliedAlertVariableNames = []string{"name", "team", "entity"}
)

type listResponse[T any] struct {
	Items []T `json:"items"`
}

type updateRequestBody struct {
	Configs firehoseConfigs `json:"configs"`
}

type resetRequestBody struct {
	To       string     `json:"to"`
	DateTime *time.Time `json:"date_time"`
}

func handleListFirehoses(client entropyv1beta1.ResourceServiceClient, shieldClient shieldv1beta1.ShieldServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		prj, err := getProject(r, shieldClient)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		rpcReq := &entropyv1beta1.ListResourcesRequest{
			Kind:    kindFirehose,
			Project: prj.GetSlug(),
		}

		rpcResp, err := client.ListResources(r.Context(), rpcReq)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		var arr []firehoseDefinition
		for _, res := range rpcResp.GetResources() {
			firehoseDef, err := mapResourceToFirehose(res, true)
			if err != nil {
				utils.WriteErr(w, err)
				return
			}
			arr = append(arr, *firehoseDef)
		}

		resp := listResponse[firehoseDefinition]{Items: arr}
		utils.WriteJSON(w, http.StatusOK, resp)
	}
}

func handleCreateFirehose(client entropyv1beta1.ResourceServiceClient, shieldClient shieldv1beta1.ShieldServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		prj, err := getProject(r, shieldClient)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		var def firehoseDefinition
		if err := json.NewDecoder(r.Body).Decode(&def); err != nil {
			utils.WriteErr(w, errors.ErrInvalid.
				WithMsgf("json body is not valid").
				WithCausef(err.Error()))
			return
		}

		res, err := mapFirehoseToResource(reqctx.From(r.Context()), def, prj)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		rpcReq := &entropyv1beta1.CreateResourceRequest{Resource: res}
		rpcResp, err := client.CreateResource(r.Context(), rpcReq)
		if err != nil {
			outErr := errors.ErrInternal

			st := status.Convert(err)
			if st.Code() == codes.AlreadyExists {
				outErr = errors.ErrConflict.WithCausef(st.Message())
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
}

func handleGetFirehose(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urn := mux.Vars(r)[pathParamURN]

		// Ensure that the URN refers to a valid firehose resource.
		def, err := getFirehoseResource(r.Context(), client, urn)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		utils.WriteJSON(w, http.StatusOK, def)
	}
}

func handleGetFirehoseHistory(entropyClient entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urn := mux.Vars(r)[pathParamURN]

		def, err := getFirehoseHistory(r.Context(), entropyClient, urn)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		utils.WriteJSON(w, http.StatusOK, def)
	}
}

func handleUpdateFirehose(client entropyv1beta1.ResourceServiceClient, shieldClient shieldv1beta1.ShieldServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathVars := mux.Vars(r)
		urn := pathVars[pathParamURN]

		prj, err := getProject(r, shieldClient)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		firehoseDef, err := getFirehoseResource(r.Context(), client, urn)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		var updReq updateRequestBody
		if err := json.NewDecoder(r.Body).Decode(&updReq); err != nil {
			utils.WriteErr(w, errors.ErrInvalid.
				WithMsgf("invalid json body").
				WithCausef(err.Error()))
			return
		}

		cfgStruct, err := updReq.Configs.toConfigStruct(prj)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		rCtx := reqctx.From(r.Context())
		labels := firehoseDef.getLabels()
		labels["updated_by"] = rCtx.UserID

		rpcReq := &entropyv1beta1.UpdateResourceRequest{
			Urn:    urn,
			Labels: labels,
			NewSpec: &entropyv1beta1.ResourceSpec{
				Configs: cfgStruct,
			},
		}

		rpcResp, err := client.UpdateResource(r.Context(), rpcReq)
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.InvalidArgument {
				utils.WriteErr(w, errors.ErrInvalid.WithCausef(st.Message()))
			} else if st.Code() == codes.NotFound {
				utils.WriteErr(w, errors.ErrNotFound.
					WithMsgf(firehoseNotFound).
					WithCausef(st.Message()))
			} else {
				utils.WriteErr(w, err)
			}
			return
		}

		firehoseDef, err = mapResourceToFirehose(rpcResp.GetResource(), false)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		utils.WriteJSON(w, http.StatusOK, firehoseDef)
	}
}

func handleDeleteFirehose(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urn := mux.Vars(r)[pathParamURN]

		// Ensure that the URN refers to a valid firehose resource.
		if _, err := getFirehoseResource(r.Context(), client, urn); err != nil {
			utils.WriteErr(w, err)
			return
		}

		_, err := client.DeleteResource(r.Context(), &entropyv1beta1.DeleteResourceRequest{Urn: urn})
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.NotFound {
				utils.WriteErr(w, errors.ErrNotFound.WithMsgf(firehoseNotFound))
				return
			}
			utils.WriteErr(w, err)
			return
		}

		utils.WriteJSON(w, http.StatusNoContent, nil)
	}
}

func handleResetFirehose(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urn := mux.Vars(r)[pathParamURN]

		// Ensure that the URN refers to a valid firehose resource.
		firehoseDef, err := getFirehoseResource(r.Context(), client, urn)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		var reqBody resetRequestBody
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			utils.WriteErr(w, errors.ErrInvalid.WithMsgf("invalid json body").WithCausef(err.Error()))
			return
		}

		paramsStruct, err := toProtobufStruct(reqBody)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		rCtx := reqctx.From(r.Context())
		labels := firehoseDef.getLabels()
		labels["updated_by"] = rCtx.UserID

		rpcReq := &entropyv1beta1.ApplyActionRequest{
			Urn:    urn,
			Action: actionResetOffset,
			Params: paramsStruct,
			Labels: labels,
		}

		rpcResp, err := client.ApplyAction(r.Context(), rpcReq)
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.InvalidArgument {
				utils.WriteErr(w, errors.ErrInvalid.WithCausef(st.Message()))
			} else if st.Code() == codes.NotFound {
				utils.WriteErr(w, errors.ErrNotFound.
					WithMsgf(firehoseNotFound).
					WithCausef(st.Message()))
			} else {
				utils.WriteErr(w, err)
			}
			return
		}

		firehoseDef, err = mapResourceToFirehose(rpcResp.GetResource(), false)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		utils.WriteJSON(w, http.StatusOK, firehoseDef)
	}
}

func handleScaleFirehose(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urn := mux.Vars(r)[pathParamURN]

		firehoseDef, err := getFirehoseResource(r.Context(), client, urn)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		var reqBody struct {
			Replicas int `json:"replicas"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			utils.WriteErr(w, errors.ErrInvalid.WithMsgf("invalid json body").WithCausef(err.Error()))
			return
		}

		paramsStruct, err := toProtobufStruct(reqBody)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		rCtx := reqctx.From(r.Context())
		labels := firehoseDef.getLabels()
		labels["updated_by"] = rCtx.UserID

		rpcReq := &entropyv1beta1.ApplyActionRequest{
			Urn:    urn,
			Action: actionScale,
			Params: paramsStruct,
			Labels: labels,
		}

		rpcResp, err := client.ApplyAction(r.Context(), rpcReq)
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.InvalidArgument {
				utils.WriteErr(w, errors.ErrInvalid.WithCausef(st.Message()))
			} else if st.Code() == codes.NotFound {
				utils.WriteErr(w, errors.ErrNotFound.
					WithMsgf(firehoseNotFound).
					WithCausef(st.Message()))
			} else {
				utils.WriteErr(w, err)
			}
			return
		}

		firehoseDef, err = mapResourceToFirehose(rpcResp.GetResource(), false)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		utils.WriteJSON(w, http.StatusOK, firehoseDef)
	}
}

func handleStartOrStop(client entropyv1beta1.ResourceServiceClient, shieldClient shieldv1beta1.ShieldServiceClient, svc *alertsv1.Service, isStop bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		urn := mux.Vars(r)[pathParamURN]

		prj, err := getProject(r, shieldClient)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		// Ensure that the URN refers to a valid firehose resource.
		firehoseDef, err := getFirehoseResource(ctx, client, urn)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		var reqBody struct{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			utils.WriteErr(w, errors.ErrInvalid.WithMsgf("invalid json body").WithCausef(err.Error()))
			return
		}

		paramsStruct, err := toProtobufStruct(reqBody)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		rCtx := reqctx.From(r.Context())
		labels := firehoseDef.getLabels()
		labels["updated_by"] = rCtx.UserID

		action := actionStart
		if isStop {
			action = actionStop
		}
		rpcReq := &entropyv1beta1.ApplyActionRequest{
			Urn:    urn,
			Action: action,
			Params: paramsStruct,
			Labels: labels,
		}

		rpcResp, err := client.ApplyAction(ctx, rpcReq)
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.InvalidArgument {
				utils.WriteErr(w, errors.ErrInvalid.WithCausef(st.Message()))
			} else if st.Code() == codes.NotFound {
				utils.WriteErr(w, errors.ErrNotFound.
					WithMsgf(firehoseNotFound).
					WithCausef(st.Message()))
			} else {
				utils.WriteErr(w, err)
			}
			return
		}

		firehoseDef, err = mapResourceToFirehose(rpcResp.GetResource(), false)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		if isStop {
			if err := stopAlertsForResource(ctx, firehoseDef, svc, prj); err != nil {
				utils.WriteErr(w, err)
				return
			}
		}

		utils.WriteJSON(w, http.StatusOK, firehoseDef)
	}
}

func stopAlertsForResource(ctx context.Context, firehoseDef *firehoseDefinition, svc *alertsv1.Service, prj *shieldv1beta1.Project) error {
	name, err := getFirehoseReleaseName(firehoseDef)
	if err != nil {
		return err
	}
	policy := alertsv1.Policy{
		Resource: name,
		Rules:    nil,
	}
	_, err = svc.UpsertAlertPolicy(ctx, prj.GetSlug(), policy)
	if err != nil {
		return err
	}
	return nil
}

func getFirehoseResource(ctx context.Context, client entropyv1beta1.ResourceServiceClient, firehoseURN string) (*firehoseDefinition, error) {
	resp, err := client.GetResource(ctx, &entropyv1beta1.GetResourceRequest{Urn: firehoseURN})
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			return nil, errors.ErrNotFound.
				WithMsgf(firehoseNotFound).
				WithCausef(st.Message())
		}
		return nil, err
	} else if resp.GetResource().GetKind() != kindFirehose {
		return nil, errors.ErrNotFound.WithMsgf(firehoseNotFound)
	}

	return mapResourceToFirehose(resp.GetResource(), false)
}

func getFirehoseHistory(ctx context.Context, client entropyv1beta1.ResourceServiceClient, firehoseURN string) ([]revisionDiff, error) {
	resp, err := client.GetResourceRevisions(ctx, &entropyv1beta1.GetResourceRevisionsRequest{Urn: firehoseURN})
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			return nil, errors.ErrNotFound.
				WithMsgf(firehoseNotFound).
				WithCausef(st.Message())
		}
		return nil, err
	}

	prevSpec := []byte("{}")
	var rh []revisionDiff

	for _, revision := range resp.GetRevisions() {
		var rd revisionDiff
		currentSpec, err := protojson.MarshalOptions{
			UseProtoNames: true,
		}.Marshal(revision.GetSpec())
		if err != nil {
			return nil, err
		}

		specDiff, err := jsonDiff(prevSpec, currentSpec)
		if err != nil {
			return nil, err
		}

		rd.Labels = revision.GetLabels()
		rd.Diff = json.RawMessage(specDiff)
		rd.UpdatedAt = revision.GetCreatedAt().AsTime()
		rh = append(rh, rd)
		prevSpec = currentSpec
	}

	return rh, nil
}

func handleGetFirehoseLogs(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			utils.WriteErr(w, errors.ErrInternal)
			return
		}

		urn := mux.Vars(r)[pathParamURN]
		queryParams := r.URL.Query()

		filters := map[string]string{}
		for _, filterKey := range firehoseLogFilterKeys {
			if queryParams.Has(filterKey) {
				filters[filterKey] = queryParams.Get(filterKey)
			}
		}

		getLogReq := &entropyv1beta1.GetLogRequest{
			Urn:    urn,
			Filter: filters,
		}

		logClient, err := client.GetLog(r.Context(), getLogReq)
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.NotFound {
				utils.WriteErr(w, errors.ErrNotFound)
			} else {
				utils.WriteErr(w, err)
			}
			return
		}

		w.Header().Set("Transfer-Encoding", "chunked")

		for {
			getLogRes, err := logClient.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					flusher.Flush()
					return
				}
				st := status.Convert(err)
				if st.Code() == codes.NotFound {
					utils.WriteErr(w, errors.ErrNotFound)
				} else {
					utils.WriteErr(w, err)
				}
				return
			}
			chunk := getLogRes.GetChunk()
			logChunk, err := protojson.Marshal(chunk)
			if err != nil {
				utils.WriteErr(w, err)
				return
			}

			utils.WriteLn(w, http.StatusOK, logChunk)
			flusher.Flush()
		}
	}
}

func handleUpgradeFirehose(client entropyv1beta1.ResourceServiceClient, shieldClient shieldv1beta1.ShieldServiceClient,
	latestFirehoseVersion string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathVars := mux.Vars(r)
		urn := pathVars[pathParamURN]

		prj, err := getProject(r, shieldClient)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		// Ensure that the URN refers to a valid firehose resource.
		cur, err := getFirehoseResource(r.Context(), client, urn)
		if err != nil {
			utils.WriteErr(w, err)
			return
		} else if cur.Configs.Version == latestFirehoseVersion {
			utils.WriteJSON(w, http.StatusNoContent, nil)
			return
		}

		cur.Configs.Version = latestFirehoseVersion
		cfgStruct, err := cur.Configs.toConfigStruct(prj)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		rCtx := reqctx.From(r.Context())
		labels := cur.getLabels()
		labels["updated_by"] = rCtx.UserID

		rpcReq := &entropyv1beta1.UpdateResourceRequest{
			Urn:    urn,
			Labels: labels,
			NewSpec: &entropyv1beta1.ResourceSpec{
				Configs: cfgStruct,
			},
		}

		rpcResp, err := client.UpdateResource(r.Context(), rpcReq)
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.InvalidArgument {
				utils.WriteErr(w, errors.ErrInvalid.WithCausef(st.Message()))
			} else if st.Code() == codes.NotFound {
				utils.WriteErr(w, errors.ErrNotFound.
					WithMsgf(firehoseNotFound).
					WithCausef(st.Message()))
			} else {
				utils.WriteErr(w, err)
			}
			return
		}

		firehoseDef, err := mapResourceToFirehose(rpcResp.GetResource(), false)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		utils.WriteJSON(w, http.StatusOK, firehoseDef)
	}
}

func handleGetFirehoseAlertPolicies(client entropyv1beta1.ResourceServiceClient, shieldClient shieldv1beta1.ShieldServiceClient, svc *alertsv1.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		pathVars := mux.Vars(r)
		urn := pathVars[pathParamURN]

		prj, err := getProject(r, shieldClient)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		firehoseDef, err := getFirehoseResource(r.Context(), client, urn)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		name, err := getFirehoseReleaseName(firehoseDef)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		policyDef, err := svc.GetAlertPolicy(ctx, prj.GetSlug(), name)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		policyDef.Rules = removeSuppliedVariablesFromRules(policyDef.Rules, suppliedAlertVariableNames)
		utils.WriteJSON(w, http.StatusOK, policyDef)
	}
}

func handleUpsertFirehoseAlertPolicies(client entropyv1beta1.ResourceServiceClient, shieldClient shieldv1beta1.ShieldServiceClient, svc *alertsv1.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		pathVars := mux.Vars(r)
		urn := pathVars[pathParamURN]

		prj, err := getProject(r, shieldClient)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		firehoseDef, err := getFirehoseResource(r.Context(), client, urn)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		name, err := getFirehoseReleaseName(firehoseDef)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}
		team := firehoseDef.Team
		entity, err := svc.GetProjectDataSource(ctx, prj.GetSlug())
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		var policyDef alertsv1.Policy
		if err := json.NewDecoder(r.Body).Decode(&policyDef); err != nil {
			utils.WriteErr(w, errors.ErrInvalid.
				WithMsgf("request json body is not valid").
				WithCausef(err.Error()))
			return
		}

		policyDef.Rules = addSuppliedVariablesFromRules(policyDef.Rules, map[string]string{
			"team":   team,
			"name":   name,
			"entity": entity,
		})
		policyDef.Resource = name

		alertPolicy, err := svc.UpsertAlertPolicy(ctx, prj.GetSlug(), policyDef)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		utils.WriteJSON(w, http.StatusOK, alertPolicy)
	}
}

func handleListFirehoseAlerts(client entropyv1beta1.ResourceServiceClient, shieldClient shieldv1beta1.ShieldServiceClient, svc *alertsv1.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		pathVars := mux.Vars(r)
		urn := pathVars[pathParamURN]

		prj, err := getProject(r, shieldClient)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		firehoseDef, err := getFirehoseResource(r.Context(), client, urn)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		name, err := getFirehoseReleaseName(firehoseDef)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		alerts, err := svc.ListAlerts(ctx, prj.GetSlug(), name)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		resp := listResponse[alertsv1.Alert]{Items: alerts}
		utils.WriteJSON(w, http.StatusOK, resp)
	}
}

func getProject(r *http.Request, shieldClient shieldv1beta1.ShieldServiceClient) (*shieldv1beta1.Project, error) {
	projectID := strings.TrimSpace(r.Header.Get(headerProjectID))
	projectSlug := mux.Vars(r)[pathParamProjectSlug]

	if projectID == "" {
		// List everything and search by slug.
		projects, err := shieldClient.ListProjects(r.Context(), &shieldv1beta1.ListProjectsRequest{})
		if err != nil {
			return nil, err
		}
		for _, prj := range projects.GetProjects() {
			if prj.GetSlug() == projectSlug {
				return prj, nil
			}
		}
		return nil, errors.ErrNotFound
	}

	// Project ID is available. Use it to fetch the project directly.
	prj, err := shieldClient.GetProject(r.Context(), &shieldv1beta1.GetProjectRequest{Id: projectID})
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			return nil, errors.ErrNotFound
		}
		return nil, err
	} else if prj.GetProject().Slug != projectSlug {
		return nil, errors.ErrNotFound.WithCausef("projectSlug in URL does not match project of given ID")
	}
	return prj.GetProject(), nil
}

func getFirehoseReleaseName(firehoseDef *firehoseDefinition) (string, error) {
	s, ok := firehoseDef.State.Output[firehoseOutputReleaseNameKey].(string)
	if !ok {
		return "", errors.ErrInternal.WithMsgf("unable to parse firehose name")
	}
	return s, nil
}

func findInArray(a []string, f string) bool {
	for _, s := range a {
		if s == f {
			return true
		}
	}
	return false
}

func removeSuppliedVariablesFromRules(rules []alertsv1.Rule, varKeys []string) []alertsv1.Rule {
	var result []alertsv1.Rule
	for _, r := range rules {
		var finalVars []alertsv1.Variable
		for _, variable := range r.Variables {
			if !findInArray(varKeys, variable.Name) {
				finalVars = append(finalVars, variable)
			}
		}
		r.Variables = finalVars
		result = append(result, r)
	}
	return result
}

func addSuppliedVariablesFromRules(rules []alertsv1.Rule, vars map[string]string) []alertsv1.Rule {
	rules = removeSuppliedVariablesFromRules(rules, maps.Keys(vars))
	var suppliedVars []alertsv1.Variable
	for k, v := range vars {
		suppliedVars = append(suppliedVars, alertsv1.Variable{
			Name:  k,
			Value: v,
		})
	}
	var result []alertsv1.Rule
	for _, rule := range rules {
		rule.Variables = append(rule.Variables, suppliedVars...)
		result = append(result, rule)
	}
	return result
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
