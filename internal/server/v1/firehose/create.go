package firehose

import (
	"encoding/json"
	"net/http"

	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"

	"github.com/odpf/dex/internal/server/utils"
)

func createFirehose(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var def firehoseDefinition
		if err := json.NewDecoder(r.Body).Decode(&def); err != nil {
			// TODO: handle error
			return
		}

		res, err := mapFirehoseToResource(def)
		if err != nil {
			// TODO: handle error
			return
		}

		rpcReq := &entropyv1beta1.CreateResourceRequest{Resource: res}
		rpcResp, err := client.CreateResource(r.Context(), rpcReq)
		if err != nil {
			// TODO: handle error.
			return
		}

		createdFirehose, err := mapResourceToFirehose(rpcResp.GetResource())
		if err != nil {
			// TODO: handle error.
			return
		}

		utils.WriteJSON(w, http.StatusCreated, createdFirehose)
	}
}
