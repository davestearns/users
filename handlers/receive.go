package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/davestearns/userservice/models"
)

func receive(r *http.Request, target interface{}) error {
	ctype := r.Header.Get(headerContentType)
	if !strings.HasPrefix(ctype, contentTypeJSON) {
		return fmt.Errorf("Incorrect content type: expected %s but got %s", contentTypeJSON, ctype)
	}
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return fmt.Errorf("error decoding request body JSON: %v", err)
	}

	if v, ok := target.(models.Validator); ok {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("error validating posted model: %v", err)
		}
	}
	return nil
}
