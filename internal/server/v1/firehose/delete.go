package firehose

import (
	"net/http"

	"github.com/gorilla/mux"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"

	"github.com/odpf/dex/internal/server/utils"
)

func deleteFirehose(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rpcReq := &entropyv1beta1.DeleteResourceRequest{Urn: mux.Vars(r)[pathParamURN]}

		_, err := client.DeleteResource(r.Context(), rpcReq)
		if err != nil {
			// TODO: handle error.
			return
		}

		utils.WriteJSON(w, http.StatusNoContent, nil)
	}
}
