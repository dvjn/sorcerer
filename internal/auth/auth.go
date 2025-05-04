package auth

import (
	"net/http"

	"github.com/dvjn/sorcerer/internal/config"
)

type Auth struct {
	middleware func(next http.Handler) http.Handler
}

func NewAuth(config *config.Config) *Auth {
	middleware := noAuth

	if config.Auth.UserHeader != "" || config.Auth.GroupsHeader != "" {
		middleware = reverseProxyHeaderAuth(config.Auth.UserHeader, config.Auth.GroupsHeader, config.Auth.GroupsHeaderSep)
	}

	return &Auth{middleware}
}

func (a *Auth) Middleware(next http.Handler) http.Handler {
	return a.middleware(next)
}
