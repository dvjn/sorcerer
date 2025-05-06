package auth

import (
	"fmt"
	"net/http"

	"github.com/dvjn/sorcerer/internal/config"
)

type Auth struct {
	middleware func(next http.Handler) http.Handler
}

func NewAuth(config *config.Config) (*Auth, error) {
	switch config.Auth.Mode {
	case "none":
		return &Auth{middleware: noAuthMiddleware()}, nil
	case "proxy-header":
		return &Auth{middleware: proxyHeaderAuthMiddleware(&config.Auth.ProxyHeaderAuth)}, nil
	}

	return nil, fmt.Errorf("invalid auth mode: %s", config.Auth.Mode)
}

func (a *Auth) Middleware(next http.Handler) http.Handler {
	return a.middleware(next)
}
