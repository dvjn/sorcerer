package jwt

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/MicahParks/jwkset"
	"github.com/rs/zerolog"
)

// JWKSProvider fetches and caches JWKS public keys with TTL-based caching
type JWKSProvider struct {
	jwksURL    string
	cacheTTL   time.Duration
	keyfunc    keyfunc.Keyfunc
	mu         sync.RWMutex
	lastFetch  time.Time
	logger     *zerolog.Logger
	cancel     context.CancelFunc
}

// NewJWKSProvider creates a new JWKS provider with caching
func NewJWKSProvider(jwksURL string, cacheTTL time.Duration, logger *zerolog.Logger) (*JWKSProvider, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Use jwkset to create a JWKS storage with caching options
	storage, err := jwkset.NewStorageFromHTTP(jwksURL, jwkset.HTTPClientStorageOptions{
		RefreshInterval: cacheTTL,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create JWKS storage: %w", err)
	}

	// Create keyfunc from storage
	jwksKeyfunc, err := keyfunc.New(keyfunc.Options{
		Ctx:     ctx,
		Storage: storage,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create keyfunc: %w", err)
	}

	p := &JWKSProvider{
		jwksURL:  jwksURL,
		cacheTTL: cacheTTL,
		keyfunc:  jwksKeyfunc,
		logger:   logger,
		cancel:   cancel,
		lastFetch: time.Now(),
	}

	p.logger.Info().
		Str("jwks_url", jwksURL).
		Str("cache_ttl", cacheTTL.String()).
		Msg("jwt jwks provider initialized")

	return p, nil
}

// GetKeyfunc returns the keyfunc for JWT verification
func (p *JWKSProvider) GetKeyfunc() keyfunc.Keyfunc {
	return p.keyfunc
}

// Close cleans up resources
func (p *JWKSProvider) Close() {
	if p.cancel != nil {
		p.cancel()
	}
}
