package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog"
)

const (
	// Clock skew tolerance - allow 30 seconds of clock drift
	clockSkewTolerance = 30 * time.Second
)

// Validator validates JWT tokens using JWKS
type Validator struct {
	provider *JWKSProvider
	issuer   string
	audience string
	verifier *oidc.IDTokenVerifier
	logger   *zerolog.Logger
}

// NewValidator creates a new JWT token validator
func NewValidator(provider *JWKSProvider, issuer, audience string, logger *zerolog.Logger) (*Validator, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	v := &Validator{
		provider: provider,
		issuer:   issuer,
		audience: audience,
		logger:   logger,
	}

	// Create verifier from provider
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oidcProvider, err := provider.GetProvider(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider for verifier: %w", err)
	}

	// Configure verifier with required claims
	v.verifier = oidcProvider.Verifier(&oidc.Config{
		ClientID:          audience,
		SkipClientIDCheck: false,
		SkipIssuerCheck:   false,
		SkipExpiryCheck:   false,
		Now:               func() time.Time { return time.Now() },
	})

	v.logger.Info().
		Str("issuer", issuer).
		Str("audience", audience).
		Msg("jwt validator initialized")

	return v, nil
}

// Validate validates a JWT token and returns its claims
func (v *Validator) Validate(ctx context.Context, token string) (*oidc.IDToken, error) {
	if token == "" {
		return nil, v.newValidationError("missing token")
	}

	// Parse and verify the token using the verifier
	idToken, err := v.verifier.Verify(ctx, token)
	if err != nil {
		v.logger.Debug().
			Err(err).
			Msg("jwt verification failed")
		return nil, v.newValidationError("invalid or expired JWT token")
	}

	// Extract subject claim for logging
	var claims struct {
		Sub string `json:"sub"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return nil, v.newValidationError("failed to extract subject claim")
	}

	v.logger.Debug().
		Str("sub", claims.Sub).
		Msg("jwt validated successfully")

	return idToken, nil
}

// newValidationError creates a standardized validation error
func (v *Validator) newValidationError(message string) error {
	return fmt.Errorf("jwt validation failed: %s", message)
}
