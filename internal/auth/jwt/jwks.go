package jwt

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog"
)

// JWKSProvider fetches and caches JWKS public keys with TTL-based caching
type JWKSProvider struct {
	jwksURL    string
	cacheTTL   time.Duration
	provider   *oidc.Provider
	mu         sync.RWMutex
	lastFetch time.Time
	logger    *zerolog.Logger
}

// NewJWKSProvider creates a new JWKS provider with caching
func NewJWKSProvider(jwksURL string, cacheTTL time.Duration, logger *zerolog.Logger) (*JWKSProvider, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	p := &JWKSProvider{
		jwksURL:  jwksURL,
		cacheTTL: cacheTTL,
		logger:   logger,
	}

	// Initial fetch
	if err := p.refreshProvider(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize JWKS provider: %w", err)
	}

	p.logger.Info().
		Str("jwks_url", jwksURL).
		Str("cache_ttl", cacheTTL.String()).
		Msg("jwt jwks provider initialized")

	return p, nil
}

// GetProvider returns the OIDC provider, refreshing cache if needed
func (p *JWKSProvider) GetProvider(ctx context.Context) (*oidc.Provider, error) {
	p.mu.RLock()
	needsRefresh := time.Since(p.lastFetch) > p.cacheTTL
	p.mu.RUnlock()

	if needsRefresh {
		if err := p.refreshProvider(ctx); err != nil {
			// If refresh fails, use cached provider if available
			p.mu.RLock()
			defer p.mu.RUnlock()
			if p.provider == nil {
				return nil, fmt.Errorf("no cached provider and refresh failed: %w", err)
			}
			p.logger.Warn().
				Err(err).
				Str("jwks_url", p.jwksURL).
				Msg("jwks refresh failed, using cached keys")
		}
	}

	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.provider, nil
}

// refreshProvider fetches the JWKS provider from the remote URL
func (p *JWKSProvider) refreshProvider(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	provider, err := oidc.NewProvider(ctx, p.jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS from %s: %w", p.jwksURL, err)
	}

	p.provider = provider
	p.lastFetch = time.Now()

	p.logger.Debug().
		Str("jwks_url", p.jwksURL).
		Msg("jwks provider refreshed")

	return nil
}
