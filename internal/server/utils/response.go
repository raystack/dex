package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// WriteJSON writes 'v' to response-writer in JSON format.
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error: failed to write 'v' JSON: %v", err)
	}
}
