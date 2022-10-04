package firehose

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"

	"github.com/odpf/dex/internal/server/utils"
)

type updateRequestBody struct {
	Configs firehoseConfigs `json:"configs"`
}

func updateFirehose(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var updReq updateRequestBody
		if err := json.NewDecoder(r.Body).Decode(&updReq); err != nil {
			// TODO: handle error
			return
		}

		cfgStruct, err := updReq.Configs.toConfigStruct()
		if err != nil {
			// TODO: handle error
			return
		}

		rpcReq := &entropyv1beta1.UpdateResourceRequest{
			// TODO: how to ensure URN refers to a firehose only?
			Urn: mux.Vars(r)[pathParamURN],
			NewSpec: &entropyv1beta1.ResourceSpec{
				Configs: cfgStruct,
			},
		}

		rpcResp, err := client.UpdateResource(r.Context(), rpcReq)
		if err != nil {
			// TODO: handle error.
			return
		}

		firehoseDef, err := mapResourceToFirehose(rpcResp.GetResource())
		if err != nil {
			// TODO: handle error.
			return
		}

		utils.WriteJSON(w, http.StatusOK, firehoseDef)
	}
}
