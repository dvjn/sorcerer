package auth

import (
	"fmt"
	"net/http"

	"github.com/dvjn/sorcerer/internal/auth/no_auth"
	"github.com/dvjn/sorcerer/internal/auth/proxy_header_auth"
	"github.com/dvjn/sorcerer/internal/config"
)

type Auth interface {
	Middleware(next http.Handler) http.Handler
}

func New(c *config.Config) (Auth, error) {
	switch c.Auth.Mode {
	case "none":
		return no_auth.New(), nil
	case "proxy-header":
		return proxy_header_auth.New(&c.Auth.ProxyHeaderAuth), nil
	}

	return nil, fmt.Errorf("invalid auth mode: %s", c.Auth.Mode)
}
