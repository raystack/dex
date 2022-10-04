package firehose

import (
	"net/http"

	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
)

func updateFirehose(client entropyv1beta1.ResourceServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
