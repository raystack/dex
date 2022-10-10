package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/odpf/dex/pkg/errors"
)

// WriteJSON writes 'v' to response-writer in JSON format.
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if status != http.StatusNoContent {
		if err := json.NewEncoder(w).Encode(v); err != nil {
			log.Printf("error: failed to write 'v' JSON: %v", err)
		}
	}
}

// WriteErr interprets the given error as one of the errors defined
// in errors package and writes the error response.
func WriteErr(w http.ResponseWriter, err error) {
	e := errors.E(err)
	WriteJSON(w, e.HTTPStatus(), e)
}
