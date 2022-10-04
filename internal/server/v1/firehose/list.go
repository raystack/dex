package firehose

import (
	"net/http"

	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"

	"github.com/odpf/dex/internal/server/utils"
)

type listResponse[T any] struct {
	Items []T `json:"items"`
}

func listFirehoses(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rpcReq := &entropyv1beta1.ListResourcesRequest{Kind: kindFirehose}

		rpcResp, err := client.ListResources(r.Context(), rpcReq)
		if err != nil {
			// TODO: handle error.
			return
		}

		var arr []firehoseDefinition
		for _, res := range rpcResp.GetResources() {
			firehoseDef, err := mapResourceToFirehose(res)
			if err != nil {
				// TODO: handle error
			}
			arr = append(arr, *firehoseDef)
		}

		resp := listResponse[firehoseDefinition]{Items: arr}
		utils.WriteJSON(w, http.StatusOK, resp)
	}
}
