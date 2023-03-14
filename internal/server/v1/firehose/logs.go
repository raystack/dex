package firehose

import (
	"io"
	"log"
	"net/http"

	entropyv1beta1 "buf.build/gen/go/gotocompany/proton/protocolbuffers/go/gotocompany/entropy/v1beta1"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/goto/dex/internal/server/utils"
	"github.com/goto/dex/pkg/errors"
)

var firehoseLogFilterKeys = []string{
	"pod", "container", "since_seconds", "tail_lines",
	"follow", "previous", "timestamps",
}

func (api *firehoseAPI) handleStreamLog(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		utils.WriteErr(w, errors.ErrInternal.WithMsgf("streaming not supported by client"))
		return
	}

	urn := chi.URLParam(r, pathParamURN)
	query := r.URL.Query()

	filters := map[string]string{}
	for _, filterKey := range firehoseLogFilterKeys {
		if query.Has(filterKey) {
			filters[filterKey] = query.Get(filterKey)
		}
	}

	rpcReq := &entropyv1beta1.GetLogRequest{
		Urn:    urn,
		Filter: filters,
	}

	logClient, err := api.Entropy.GetLog(r.Context(), rpcReq)
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			utils.WriteErr(w, errors.ErrNotFound)
		} else {
			utils.WriteErr(w, err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Transfer-Encoding", "chunked")

	statusSent := false
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

		logChunk, err := protojson.Marshal(getLogRes.GetChunk())
		if err != nil {
			utils.WriteErr(w, err)
			return
		}

		if !statusSent {
			w.WriteHeader(http.StatusOK)
			statusSent = true
		}

		writeLine(w, logChunk)
		flusher.Flush()
	}
}

func writeLine(w http.ResponseWriter, b []byte) {
	b = append(b, '\n')
	if _, err := w.Write(b); err != nil {
		log.Printf("error: failed to write line: %v", err)
	}
}
