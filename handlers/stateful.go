package handlers

import (
	"net/http"
)

//StatefulHandlerFunc is an HTTP handler function that requires session state
type StatefulHandlerFunc func(http.ResponseWriter, *http.Request, *SessionState)

//EnsureSession is an adapter that converts a StatefulHandlerFunc into an http.HandlerFunc
func (c *Config) EnsureSession(handlerFunc StatefulHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionState := &SessionState{}
		if _, err := c.SessionManager.GetState(r, sessionState); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		handlerFunc(w, r, sessionState)
	}
}
