package no_auth

import (
	"net/http"

	"github.com/dvjn/sorcerer/internal/config"
	"github.com/go-chi/chi/v5"
)

type NoAuth struct {
	config *config.NoAuthConfig
}

func New(config *config.NoAuthConfig) *NoAuth {
	return &NoAuth{config: config}
}

func (a *NoAuth) Router() *chi.Mux {
	r := chi.NewRouter()
	return r
}

func (a *NoAuth) DistributionMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return next
	}
}
