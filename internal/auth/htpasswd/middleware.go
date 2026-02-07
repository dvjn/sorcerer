package htpasswd

import (
	"context"
	"net/http"
)

type contextKey string

const userContextKey contextKey = "user"

func (a *HtpasswdAuth) DistributionMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health checks and metrics
			if a.shouldSkipAuth(r) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract Basic Auth credentials
			username, password, ok := r.BasicAuth()
			if !ok {
				a.logger.Debug().
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Msg("missing or invalid basic auth credentials")
				a.challenge(w)
				return
			}

			// Validate credentials using the library
			if !a.Match(username, password) {
				a.logger.Warn().
					Str("username", username).
					Str("path", r.URL.Path).
					Msg("authentication failed")
				a.challenge(w)
				return
			}

			// Log successful authentication
			a.logger.Info().
				Str("username", username).
				Str("path", r.URL.Path).
				Str("method", r.Method).
				Msg("user authenticated successfully")

			// Set user context and continue
			ctx := context.WithValue(r.Context(), userContextKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *HtpasswdAuth) challenge(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Sorcerer OCI Registry"`)
	w.WriteHeader(http.StatusUnauthorized)
}

func (a *HtpasswdAuth) shouldSkipAuth(r *http.Request) bool {
	// Skip auth for certain endpoints
	skipPaths := []string{
		"/health",
		"/metrics",
	}

	// Allow unauthenticated access only to the base /v2/ endpoint for discovery
	if r.URL.Path == "/v2/" && r.Method == "GET" {
		return true
	}

	for _, path := range skipPaths {
		if r.URL.Path == path {
			return true
		}
	}
	return false
}

func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(userContextKey).(string)
	return username, ok
}

func (a *HtpasswdAuth) handleError(w http.ResponseWriter, err error, context string) {
	// Always return generic "unauthorized" - don't reveal user existence
	a.logger.Error().
		Err(err).
		Str("context", context).
		Msg("authentication error")
	a.challenge(w)
}
