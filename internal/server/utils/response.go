package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/goto/dex/pkg/errors"
)

// ListResponse can be used to write list of items to response.
// This format is helpful in enabling pagination.
type ListResponse[T any] struct {
	Items []T `json:"items"`
}

func ReadJSON(r *http.Request, into any) error {
	if err := json.NewDecoder(r.Body).Decode(into); err != nil {
		return errors.ErrInvalid.
			WithMsgf("json body is not valid").WithCausef(err.Error())
	}
	return nil
}

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

// WriteErr interprets the given error as one of the errors defined
// in errors package and writes the error response.
func WriteErrMsg(w http.ResponseWriter, statusCode int, message string) {
	err := errors.Error{
		Message: message,
		Status:  statusCode,
	}
	WriteJSON(w, err.HTTPStatus(), err)
}
