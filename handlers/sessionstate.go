package handlers

import (
	"net/http"
	"time"

	"github.com/davestearns/userservice/models/users"
)

//SessionState represents the state of a session
type SessionState struct {
	Began        time.Time
	ClientIPPath string
	User         *users.User
}

//NewSessionState constructs a new SessionState
func NewSessionState(r *http.Request, user *users.User) *SessionState {
	var ipPath string
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if len(forwardedFor) > 0 {
		ipPath = forwardedFor + " " + r.RemoteAddr
	} else {
		ipPath = r.RemoteAddr
	}
	return &SessionState{
		Began:        time.Now(),
		ClientIPPath: ipPath,
		User:         user,
	}
}
