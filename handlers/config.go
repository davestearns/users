package handlers

import (
	"github.com/davestearns/sessions"
	"github.com/davestearns/userservice/models/users"
)

//Config holds the global configuration values for handlers
type Config struct {
	SessionManager sessions.Manager
	UserStore      users.Store
}
