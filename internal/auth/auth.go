package auth

import (
	"fmt"
	"net/http"

	"github.com/dvjn/sorcerer/internal/auth/htpasswd"
	"github.com/dvjn/sorcerer/internal/auth/jwt"
	"github.com/dvjn/sorcerer/internal/auth/no_auth"
	"github.com/dvjn/sorcerer/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

type Auth interface {
	Router() *chi.Mux
	DistributionMiddleware() func(http.Handler) http.Handler
}

func New(c *config.AuthConfig, logger *zerolog.Logger) (Auth, error) {
	switch c.Mode {
	case config.AuthModeNone:
		return no_auth.New(&c.NoAuth), nil
	case config.AuthModeHtpasswd:
		return htpasswd.NewHtpasswdAuth(&c.Htpasswd, logger)
	case config.AuthModeJWT:
		return jwt.NewJWTAuth(&c.JWT, logger)
	default:
		return nil, fmt.Errorf("unknown auth mode: %s", c.Mode)
	}
}
