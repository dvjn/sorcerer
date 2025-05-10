package auth

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/dvjn/sorcerer/internal/auth/no_auth"
	"github.com/dvjn/sorcerer/internal/config"
	"github.com/go-chi/chi/v5"
)

type Auth interface {
	Router() *chi.Mux
	DistributionMiddleware() func(http.Handler) http.Handler
}

func New(c *config.AuthConfig) (Auth, error) {
	switch c.Mode {
	case config.AuthModeNone:
		return no_auth.New(&c.NoAuth), nil
	default:
		return nil, fmt.Errorf("unknown auth mode: %s", c.Mode)
	}
}
