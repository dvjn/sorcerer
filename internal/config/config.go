package config

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

const (
	AUTH_MODE_NONE         = "none"
	AUTH_MODE_PROXY_HEADER = "proxy-header"
)

type ProxyHeaderAuthConfig struct {
	UserHeaderName        string `koanf:"user_header_name"`
	GroupsHeaderName      string `koanf:"groups_header_name"`
	GroupsHeaderSeparator string `koanf:"groups_header_separator"`
}

type AuthConfig struct {
	Mode        string                `koanf:"mode"`
	ProxyHeader ProxyHeaderAuthConfig `koanf:"proxy_header"`
}

type StoreConfig struct {
	Path string `koanf:"path"`
}

type Config struct {
	Port  int         `koanf:"port"`
	Store StoreConfig `koanf:"store"`
	Auth  AuthConfig  `koanf:"auth"`
}

func Load() (*Config, error) {
	k := koanf.New("__")
	config := Config{}

	k.Load(structs.Provider(Config{
		Port: 3000,
		Store: StoreConfig{
			Path: "data",
		},
		Auth: AuthConfig{
			Mode: AUTH_MODE_NONE,
			ProxyHeader: ProxyHeaderAuthConfig{
				GroupsHeaderSeparator: ",",
			},
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

	if c.Auth.Mode == AUTH_MODE_PROXY_HEADER && c.Auth.ProxyHeader.UserHeaderName == "" {
		errors = append(errors, fmt.Errorf("auth.proxy_header.user_header_name must be set when auth.mode is %s", AUTH_MODE_PROXY_HEADER))
	}

	return errors
}
