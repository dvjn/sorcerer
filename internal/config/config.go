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
)

type NoAuthConfig struct{}

type HtpasswdConfig struct {
	File     string `koanf:"file"`     // Path to htpasswd file
	Contents string `koanf:"contents"` // Inline htpasswd content
}

type AuthConfig struct {
	Mode     string          `koanf:"mode"`
	NoAuth   NoAuthConfig    `koanf:"no_auth"`
	Htpasswd HtpasswdConfig  `koanf:"htpasswd"`
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

	if c.Auth.Mode != AuthModeNone && c.Auth.Mode != AuthModeHtpasswd {
		errors = append(errors, fmt.Errorf("invalid auth mode: %s", c.Auth.Mode))
	}

	// Additional validation for htpasswd mode
	if c.Auth.Mode == AuthModeHtpasswd {
		if c.Auth.Htpasswd.File == "" && c.Auth.Htpasswd.Contents == "" {
			errors = append(errors, fmt.Errorf("htpasswd auth mode requires either file or contents to be specified"))
		}
	}

	return errors
}
