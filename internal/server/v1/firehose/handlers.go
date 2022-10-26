package firehose

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/odpf/dex/internal/server/reqctx"
	"github.com/odpf/dex/internal/server/utils"
	"github.com/odpf/dex/pkg/errors"
)

const firehoseNotFound = "no firehose with given URN"

var firehoseLogFilterKeys = []string{"pod", "container", "sinceSeconds", "tailLines", "follow", "previous", "timestamps"}

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

func handleListFirehoses(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rpcReq := &entropyv1beta1.ListResourcesRequest{Kind: kindFirehose}

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
		projectID := r.Header.Get(headerProjectID)

		prj, err := shieldClient.GetProject(r.Context(), &shieldv1beta1.GetProjectRequest{Id: projectID})
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.NotFound {
				utils.WriteErr(w, errors.ErrNotFound)
			} else {
				utils.WriteErr(w, err)
			}
			return
		}

		var def firehoseDefinition
		if err := json.NewDecoder(r.Body).Decode(&def); err != nil {
			utils.WriteErr(w, errors.ErrInvalid.
				WithMsgf("json body is not valid").
				WithCausef(err.Error()))
			return
		}

		res, err := mapFirehoseToResource(reqctx.From(r.Context()), def, prj.GetProject())
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

func handleUpdateFirehose(client entropyv1beta1.ResourceServiceClient, shieldClient shieldv1beta1.ShieldServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathVars := mux.Vars(r)
		projectID := r.Header.Get(headerProjectID)
		urn := pathVars[pathParamURN]

		getProjectResponse, err := shieldClient.GetProject(r.Context(), &shieldv1beta1.GetProjectRequest{Id: projectID})
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.NotFound {
				utils.WriteErr(w, errors.ErrNotFound)
			} else {
				utils.WriteErr(w, err)
			}
			return
		}

		// Ensure that the URN refers to a valid firehose resource.
		if _, err := getFirehoseResource(r.Context(), client, urn); err != nil {
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

		cfgStruct, err := updReq.Configs.toConfigStruct(getProjectResponse.GetProject())
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		rpcReq := &entropyv1beta1.UpdateResourceRequest{
			Urn:    urn,
			Labels: map[string]string{}, // TODO: merge shield labels with current value.
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
		if _, err := getFirehoseResource(r.Context(), client, urn); err != nil {
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

		rpcReq := &entropyv1beta1.ApplyActionRequest{
			Urn:    urn,
			Action: actionResetOffset,
			Params: paramsStruct,
			Labels: map[string]string{}, // TODO: shield labels.
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

		firehoseDef, err := mapResourceToFirehose(rpcResp.GetResource(), false)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		utils.WriteJSON(w, http.StatusOK, firehoseDef)
	}
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
				fmt.Println("err ", err)
				utils.WriteErr(w, err)
				return
			}

			utils.WriteJSON(w, http.StatusOK, json.RawMessage(logChunk))
			flusher.Flush()
		}
	}
}
