package firehose

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	shieldv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/odpf/dex/internal/server/utils"
	"github.com/odpf/dex/pkg/errors"
)

const firehoseNotFound = "no firehose with given URN"

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

func listFirehoses(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
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

func createFirehose(client entropyv1beta1.ResourceServiceClient, shieldClient shieldv1beta1.ShieldServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := mux.Vars(r)[pathParamProjectID]
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

		res, err := mapFirehoseToResource(def, prj.GetProject())
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

func getFirehose(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urn := mux.Vars(r)[pathParamURN]

		getResReq := &entropyv1beta1.GetResourceRequest{Urn: urn}
		resp, err := client.GetResource(r.Context(), getResReq)
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.NotFound {
				utils.WriteErr(w, errors.ErrNotFound)
			} else {
				utils.WriteErr(w, err)
			}
			return
		} else if resp.GetResource() == nil {
			utils.WriteErr(w, errors.ErrNotFound)
			return
		}

		res := resp.GetResource()
		if res.GetKind() != kindFirehose {
			utils.WriteErr(w, errors.ErrNotFound)
			return
		}

		def, err := mapResourceToFirehose(res, false)
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		utils.WriteJSON(w, http.StatusOK, def)
	}
}

func updateFirehose(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urn := mux.Vars(r)[pathParamURN]
		curRes, err := client.GetResource(r.Context(), &entropyv1beta1.GetResourceRequest{Urn: urn})
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.NotFound {
				utils.WriteErr(w, errors.ErrNotFound.
					WithMsgf(firehoseNotFound).
					WithCausef(st.Message()))
			} else {
				utils.WriteErr(w, err)
			}
			return
		} else if curRes.GetResource().GetKind() != kindFirehose {
			utils.WriteErr(w, errors.ErrNotFound.WithMsgf(firehoseNotFound))
			return
		}

		var updReq updateRequestBody
		if err := json.NewDecoder(r.Body).Decode(&updReq); err != nil {
			utils.WriteErr(w, errors.ErrInvalid.
				WithMsgf("invalid json body").
				WithCausef(err.Error()))
			return
		}

		cfgStruct, err := updReq.Configs.toConfigStruct()
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

func deleteFirehose(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rpcReq := &entropyv1beta1.DeleteResourceRequest{Urn: mux.Vars(r)[pathParamURN]}

		_, err := client.DeleteResource(r.Context(), rpcReq)
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

func resetOffset(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urn := mux.Vars(r)[pathParamURN]
		curRes, err := client.GetResource(r.Context(), &entropyv1beta1.GetResourceRequest{Urn: urn})
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.NotFound {
				utils.WriteErr(w, errors.ErrNotFound.
					WithMsgf(firehoseNotFound).
					WithCausef(st.Message()))
			} else {
				utils.WriteErr(w, err)
			}
			return
		} else if curRes.GetResource().GetKind() != kindFirehose {
			utils.WriteErr(w, errors.ErrNotFound.WithMsgf(firehoseNotFound))
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
			Urn:    curRes.GetResource().GetUrn(),
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
