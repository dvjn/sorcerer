package config

import (
	"flag"
	"os"
)

type Config struct {
	StoragePath string
	Port        string
}

type param struct {
	field      *string
	flagName   string
	envName    string
	defaultVal string
	usage      string
}

func LoadConfig() *Config {
	var config Config

	params := []param{
		{
			field:      &config.StoragePath,
			flagName:   "storage-path",
			envName:    "STORAGE_PATH",
			defaultVal: "data",
			usage:      "The root directory for the storage",
		},
		{
			field:      &config.Port,
			flagName:   "port",
			envName:    "PORT",
			defaultVal: "3000",
			usage:      "The port to run the server on",
		},
	}

	loadParameters(params)

	return &config
}

func loadParameters(params []param) {
	for _, p := range params {
		flag.StringVar(p.field, p.flagName, p.defaultVal, p.usage)
	}
	flag.Parse()

	for _, p := range params {
		if envVal := os.Getenv(p.envName); envVal != "" {
			*p.field = envVal
		}
	}
}
