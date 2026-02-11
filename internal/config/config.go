package config

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
)

type LogConfig struct {
	Level string `koanf:"level"`
}

type ServerConfig struct {
	Port int `koanf:"port"`
}

const (
	AuthModeNone     = "none"
	AuthModeHtpasswd = "htpasswd"
	AuthModeJWT      = "jwt"
)

type NoAuthConfig struct{}

type HtpasswdConfig struct {
	File     string `koanf:"file"`     // Path to htpasswd file
	Contents string `koanf:"contents"` // Inline htpasswd content
}

type JWTConfig struct {
	JWKSURL  string `koanf:"jwks_url"`  // URL to JWKS endpoint
	Issuer   string `koanf:"issuer"`    // Expected token issuer
	Audience string `koanf:"audience"`  // Expected token audience
	CacheTTL int    `koanf:"cache_ttl"` // JWKS cache duration in seconds, default: 300
}

type AuthConfig struct {
	Mode     string          `koanf:"mode"`
	NoAuth   NoAuthConfig    `koanf:"no_auth"`
	Htpasswd HtpasswdConfig  `koanf:"htpasswd"`
	JWT      JWTConfig       `koanf:"jwt"`
}

type StoreConfig struct {
	Path string `koanf:"path"`
}

type Config struct {
	Log    LogConfig    `koanf:"log"`
	Server ServerConfig `koanf:"server"`
	Auth   AuthConfig   `koanf:"auth"`
	Store  StoreConfig  `koanf:"store"`
}

func Load() (*Config, error) {
	k := koanf.New("__")
	config := Config{}

	k.Load(structs.Provider(Config{
		Log: LogConfig{
			Level: zerolog.LevelInfoValue,
		},
		Server: ServerConfig{
			Port: 3000,
		},
		Auth: AuthConfig{
			Mode:     AuthModeNone,
			NoAuth:   NoAuthConfig{},
			Htpasswd: HtpasswdConfig{},
		},
		Store: StoreConfig{
			Path: "data",
		},
	}, "koanf"), nil)

	k.Load(env.Provider("", "__", func(s string) string {
		return strings.ToLower(s)
	}), nil)

	if err := k.Unmarshal("", &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Validate() []error {
	errors := []error{}

	if c.Auth.Mode != AuthModeNone && c.Auth.Mode != AuthModeHtpasswd && c.Auth.Mode != AuthModeJWT {
		errors = append(errors, fmt.Errorf("invalid auth mode: %s", c.Auth.Mode))
	}

	// Additional validation for htpasswd mode
	if c.Auth.Mode == AuthModeHtpasswd {
		if c.Auth.Htpasswd.File == "" && c.Auth.Htpasswd.Contents == "" {
			errors = append(errors, fmt.Errorf("htpasswd auth mode requires either file or contents to be specified"))
		}
	}

	// Additional validation for jwt mode
	if c.Auth.Mode == AuthModeJWT {
		if c.Auth.JWT.JWKSURL == "" {
			errors = append(errors, fmt.Errorf("jwt auth mode requires jwks_url to be specified"))
		}
		if c.Auth.JWT.Issuer == "" {
			errors = append(errors, fmt.Errorf("jwt auth mode requires issuer to be specified"))
		}
		if c.Auth.JWT.Audience == "" {
			errors = append(errors, fmt.Errorf("jwt auth mode requires audience to be specified"))
		}
		if c.Auth.JWT.CacheTTL == 0 {
			c.Auth.JWT.CacheTTL = 300 // Default TTL: 5 minutes
		}
	}

	return errors
}
