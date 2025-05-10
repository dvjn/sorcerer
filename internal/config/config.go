package config

import (
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

type StoreConfig struct {
	Path string `koanf:"path"`
}

type Config struct {
	Port  int         `koanf:"port"`
	Store StoreConfig `koanf:"store"`
}

func Load() (*Config, error) {
	k := koanf.New("__")
	config := Config{}

	k.Load(structs.Provider(Config{
		Port: 3000,
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

	return errors
}
