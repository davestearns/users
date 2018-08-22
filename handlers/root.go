package handlers

import (
	"net/http"
)

//RootHandler handles requests for the root resource
func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("welcome to the user service"))
}
