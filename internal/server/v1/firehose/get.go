package firehose

import (
	"net/http"

	"github.com/gorilla/mux"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"

	"github.com/odpf/dex/internal/server/utils"
)

func getFirehose(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urn := mux.Vars(r)[pathParamURN]

		getResReq := &entropyv1beta1.GetResourceRequest{Urn: urn}
		resp, err := client.GetResource(r.Context(), getResReq)
		if err != nil || resp.GetResource() == nil {
			// TODO: error handling.
			return
		}

		res := resp.GetResource()
		if res == nil || res.GetKind() != kindFirehose {
			// TODO: treat it as not found.
			return
		}

		def, err := mapResourceToFirehose(res)
		if err != nil {
			// TODO: error handling.
			return
		}

		utils.WriteJSON(w, http.StatusOK, def)
	}
}
