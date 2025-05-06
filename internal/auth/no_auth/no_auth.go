package no_auth

import "net/http"

type NoAuth struct{}

func New() *NoAuth {
	return &NoAuth{}
}

func (a *NoAuth) Middleware(next http.Handler) http.Handler {
	return next
}
