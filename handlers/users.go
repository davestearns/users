package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/davestearns/userservice/models/users"
)

//UsersHandler handles requests for the /users resource
func (c *Config) UsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		//sign-up
		newUser := &users.NewUser{}
		if err := receive(r, newUser); err != nil {
			http.Error(w, fmt.Sprintf("error receiving posted user: %v", err), http.StatusBadRequest)
			return
		}
		existingUser, err := c.UserStore.Get(newUser.UserName)
		if err != nil {
			http.Error(w, fmt.Sprintf("error checking for existing user with name '%s': %v", newUser.UserName, err), http.StatusInternalServerError)
			return
		}
		if existingUser != nil {
			http.Error(w, fmt.Sprintf("sorry, but the user name '%s' is already taken", newUser.UserName), http.StatusInternalServerError)
			return
		}

		user, err := newUser.ToUser()
		if err != nil {
			http.Error(w, fmt.Sprintf("error validating new user: %v", err), http.StatusBadRequest)
			return
		}
		if err := c.UserStore.Insert(user); err != nil {
			http.Error(w, fmt.Sprintf("error inserting new user into database: %v", err), http.StatusInternalServerError)
			return
		}
		if _, err := c.SessionManager.BeginSession(w, NewSessionState(r, user)); err != nil {
			http.Error(w, fmt.Sprintf("error begining new session: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Add(headerLocation, "/users/"+url.PathEscape(user.UserName))
		respond(w, user, http.StatusCreated)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

//SpecificUserHandler handles requests for the /users/<username> resource
func (c *Config) SpecificUserHandler(w http.ResponseWriter, r *http.Request, sessionState *SessionState) {
	userName := path.Base(r.URL.Path)
	if userName == "me" {
		//optimization: if GET /users/me, respond with currently authenticated user
		if r.Method == http.MethodGet {
			respond(w, sessionState.User, http.StatusOK)
			return
		}
		userName = sessionState.User.UserName
	}

	switch r.Method {
	case http.MethodGet:
		//can read any user profile
		user, err := c.UserStore.Get(userName)
		if err != nil {
			http.Error(w, fmt.Sprintf("error getting user from database: %v", err), http.StatusInternalServerError)
			return
		}
		respond(w, user, http.StatusOK)

	case http.MethodPatch:
		//may update only your own profile
		if userName != sessionState.User.UserName {
			http.Error(w, "you may not update profiles of other users", http.StatusForbidden)
			return
		}
		updates := &users.Updates{}
		if err := receive(r, updates); err != nil {
			http.Error(w, fmt.Sprintf("error receiving posted updates: %v", err), http.StatusBadRequest)
			return
		}
		user, err := c.UserStore.Update(userName, updates)
		if err != nil {
			http.Error(w, fmt.Sprintf("error updating user profile: %v", err), http.StatusInternalServerError)
			return
		}
		respond(w, user, http.StatusOK)

	case http.MethodDelete:
		//may delete only your own profile
		if userName != sessionState.User.UserName {
			http.Error(w, "you may not delete profiles of other users", http.StatusForbidden)
			return
		}
		if err := c.UserStore.Delete(userName); err != nil {
			http.Error(w, fmt.Sprintf("error deleting user profile: %v", err), http.StatusInternalServerError)
			return
		}
		c.SessionManager.EndSession(r)
		w.Write([]byte("account deleted"))

	default:
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}
}
