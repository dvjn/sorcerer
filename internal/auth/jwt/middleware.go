package jwt

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dvjn/sorcerer/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

type contextKey string

const userContextKey contextKey = "user"

type authError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

type authErrorResponse struct {
	Errors []authError `json:"errors"`
}

// JWTAuth implements the Auth interface for JWT-based authentication
type JWTAuth struct {
	validator *Validator
	logger    *zerolog.Logger
}

// NewJWTAuth creates a new JWT authentication handler
func NewJWTAuth(jwtConfig *config.JWTConfig, logger *zerolog.Logger) (*JWTAuth, error) {
	if logger == nil {
		return nil, loggerRequiredError()
	}

	// Create JWKS provider
	cacheTTLDuration := time.Duration(jwtConfig.CacheTTL) * time.Second
	provider, err := NewJWKSProvider(jwtConfig.JWKSURL, cacheTTLDuration, logger)
	if err != nil {
		return nil, err
	}

	// Create validator
	validator, err := NewValidator(provider, jwtConfig.Issuer, jwtConfig.Audience, logger)
	if err != nil {
		return nil, err
	}

	auth := &JWTAuth{
		validator: validator,
		logger:    logger,
	}

	auth.logger.Info().
		Str("auth_type", "jwt").
		Str("issuer", jwtConfig.Issuer).
		Str("audience", jwtConfig.Audience).
		Msg("jwt authentication initialized")

	return auth, nil
}

func loggerRequiredError() error {
	return &loggerErr{}
}

type loggerErr struct{}

func (e *loggerErr) Error() string {
	return "logger is required"
}

func (e *loggerErr) Is(target error) bool {
	_, ok := target.(*loggerErr)
	return ok
}

// Router returns the auth router (no additional routes needed for JWT)
func (a *JWTAuth) Router() *chi.Mux {
	r := chi.NewRouter()
	return r
}

// DistributionMiddleware returns middleware for JWT authentication
func (a *JWTAuth) DistributionMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if a.shouldSkipAuth(r) {
				next.ServeHTTP(w, r)
				return
			}

			username, err := a.authenticate(r)
			if err != nil {
				a.logger.Debug().
					Err(err).
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Msg("authentication failed")
				a.writeUnauthorized(w, err)
				return
			}

			a.logger.Info().
				Str("username", username).
				Str("path", r.URL.Path).
				Str("method", r.Method).
				Msg("user authenticated successfully")

			ctx := context.WithValue(r.Context(), userContextKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// authenticate extracts and validates the JWT token
func (a *JWTAuth) authenticate(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", &authErr{message: "missing authorization header"}
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", &authErr{message: "invalid authorization header format, expected Bearer token"}
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", &authErr{message: "empty bearer token"}
	}

	ctx := r.Context()
	claims, err := a.validator.Validate(ctx, token)
	if err != nil {
		return "", &authErr{message: err.Error()}
	}

	return claims.Sub, nil
}

// writeUnauthorized writes a 401 response with Docker registry error format
func (a *JWTAuth) writeUnauthorized(w http.ResponseWriter, err error) {
	var message string
	if authErr, ok := err.(*authErr); ok {
		message = authErr.message
	} else if strings.Contains(err.Error(), "validation failed") {
		// Strip the prefix for cleaner error messages
		message = strings.TrimPrefix(err.Error(), "jwt validation failed: ")
	} else {
		message = "authentication required"
	}

	w.Header().Set("Content-Type", "application/vnd.docker.distribution.errors.v2+json")
	w.WriteHeader(http.StatusUnauthorized)

	response := authErrorResponse{
		Errors: []authError{
			{
				Code:    "UNAUTHORIZED",
				Message: "authentication required",
				Detail:  message,
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}

// shouldSkipAuth determines if auth should be skipped for this request
func (a *JWTAuth) shouldSkipAuth(r *http.Request) bool {
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

// GetUsernameFromContext extracts the username from the request context
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(userContextKey).(string)
	return username, ok
}

type authErr struct {
	message string
}

func (e *authErr) Error() string {
	return e.message
}
