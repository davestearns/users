package handlers

import (
	"encoding/json"
	"net/http"
)

func respond(w http.ResponseWriter, value interface{}, statusCode int) {
	w.Header().Add(headerContentType, contentTypeJSON)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(value)
}
