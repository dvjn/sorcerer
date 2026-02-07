package htpasswd

import (
	"fmt"
	"strings"

	"github.com/dvjn/sorcerer/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	htpasswdlib "github.com/tg123/go-htpasswd"
)

type HtpasswdAuth struct {
	config *config.HtpasswdConfig
	file   *htpasswdlib.File
	logger *zerolog.Logger
}

func NewHtpasswdAuth(cfg *config.HtpasswdConfig, logger *zerolog.Logger) (*HtpasswdAuth, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	auth := &HtpasswdAuth{
		config: cfg,
		logger: logger,
	}

	if err := auth.loadHtpasswdFile(); err != nil {
		return nil, fmt.Errorf("failed to load htpasswd data: %w", err)
	}

	// Log successful initialization
	auth.logger.Info().
		Str("auth_type", "htpasswd").
		Msg("htpasswd authentication initialized")

	return auth, nil
}

func (a *HtpasswdAuth) Router() *chi.Mux {
	r := chi.NewRouter()
	return r
}

func (a *HtpasswdAuth) loadHtpasswdFile() error {
	if a.config.Contents != "" {
		a.logger.Info().
			Str("source", "inline").
			Msg("loading htpasswd from inline contents")
		r := strings.NewReader(a.config.Contents)
		file, err := htpasswdlib.NewFromReader(r, htpasswdlib.DefaultSystems, nil)
		if err != nil {
			return fmt.Errorf("failed to parse htpasswd content: %w", err)
		}
		a.file = file
		return nil
	}

	if a.config.File != "" {
		a.logger.Info().
			Str("source", "file").
			Str("file", a.config.File).
			Msg("loading htpasswd from file")

		// Load using the library
		file, err := htpasswdlib.New(a.config.File, htpasswdlib.DefaultSystems, nil)
		if err != nil {
			return fmt.Errorf("failed to load htpasswd file %s: %w", a.config.File, err)
		}

		a.file = file
		return nil
	}

	return fmt.Errorf("neither file nor contents provided for htpasswd auth")
}

// Match checks if username and password are valid
func (a *HtpasswdAuth) Match(username, password string) bool {
	if a.file == nil {
		return false
	}
	return a.file.Match(username, password)
}
