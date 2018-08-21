package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/davestearns/userservice/models/users"
)

const invalidCredentials = "invalid credentials"

//SessionsHandler handles requests for the /sessions resource
func (c *Config) SessionsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		//sign-in
		creds := &users.Credentials{}
		if err := receive(r, creds); err != nil {
			http.Error(w, fmt.Sprintf("error receiving posted credentials: %v", err), http.StatusBadRequest)
			return
		}
		//get the user associated with that user name
		user, err := c.UserStore.Get(creds.UserName)
		if err != nil {
			log.Printf("error getting user '%s' from user store: %v", creds.UserName, err)
			users.DummyAuthenticate()
			http.Error(w, invalidCredentials, http.StatusUnauthorized)
			return
		}

		if err := user.Authenticate([]byte(creds.Password)); err != nil {
			http.Error(w, invalidCredentials, http.StatusUnauthorized)
			return
		}

		if _, err := c.SessionManager.BeginSession(w, NewSessionState(r, user)); err != nil {
			http.Error(w, fmt.Sprintf("error starting new session: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Add(headerLocation, "/sessions/mine")
		respond(w, user, http.StatusCreated)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

//SessionsMineHandler handles requests for the /sessions/mine resource
func (c *Config) SessionsMineHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		if err := c.SessionManager.EndSession(r); err != nil {
			http.Error(w, fmt.Sprintf("error ending session: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("session ended"))

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
