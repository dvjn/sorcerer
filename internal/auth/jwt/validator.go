package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
)

const (
	// Clock skew tolerance - allow 30 seconds of clock drift
	clockSkewTolerance = 30 * time.Second
)

// Claims represents the custom JWT claims we extract
type Claims struct {
	Sub string `json:"sub"`
	jwt.RegisteredClaims
}

// Validator validates JWT tokens using JWKS
type Validator struct {
	provider *JWKSProvider
	issuer   string
	audience string
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

	v.logger.Info().
		Str("issuer", issuer).
		Str("audience", audience).
		Msg("jwt validator initialized")

	return v, nil
}

// Validate validates a JWT token and returns the validated claims
func (v *Validator) Validate(ctx context.Context, tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, v.newValidationError("missing token")
	}

	// Get JWKS keyfunc for verification
	kf := v.provider.GetKeyfunc()
	if kf == nil {
		return nil, fmt.Errorf("jwks not available")
	}

	// Parse and verify the token with keyfunc
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, kf.Keyfunc, jwt.WithLeeway(clockSkewTolerance))
	if err != nil {
		v.logger.Debug().
			Err(err).
			Msg("jwt verification failed")
		return nil, v.newValidationError("invalid or expired JWT token")
	}

	// Extract claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, v.newValidationError("invalid token claims")
	}

	// Custom validation for issuer and audience (keyfunc does signature verification)
	if claims.Issuer != v.issuer {
		return nil, v.newValidationError("invalid issuer")
	}

	var audienceMatch bool
	for _, aud := range claims.Audience {
		if aud == v.audience {
			audienceMatch = true
			break
		}
	}
	if !audienceMatch {
		return nil, v.newValidationError("invalid audience")
	}

	if claims.Sub == "" {
		return nil, v.newValidationError("missing subject claim")
	}

	v.logger.Debug().
		Str("sub", claims.Sub).
		Str("iss", claims.Issuer).
		Msg("jwt validated successfully")

	return claims, nil
}

// newValidationError creates a standardized validation error
func (v *Validator) newValidationError(message string) error {
	return fmt.Errorf("jwt validation failed: %s", message)
}
