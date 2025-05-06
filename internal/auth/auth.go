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
	case config.AUTH_MODE_NONE:
		return no_auth.New(), nil
	case config.AUTH_MODE_PROXY_HEADER:
		return proxy_header_auth.New(&c.Auth.ProxyHeader), nil
	}

	return nil, fmt.Errorf("invalid auth mode: %s", c.Auth.Mode)
}
