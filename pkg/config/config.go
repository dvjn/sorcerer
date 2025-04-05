package config

import (
	"os"

	"github.com/pborman/getopt/v2"
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
			usage:      "registry data storage path",
		},
		{
			field:      &config.Port,
			flagName:   "port",
			envName:    "PORT",
			defaultVal: "3000",
			usage:      "The port to run the server on",
		},
	}

	helpFlag := getopt.BoolLong("help", 'h', "display help")

	loadParameters(params)

	if *helpFlag {
		getopt.Usage()
		os.Exit(0)
	}

	return &config
}

func loadParameters(params []param) {
	for _, p := range params {
		*p.field = p.defaultVal
	}

	for _, p := range params {
		getopt.FlagLong(p.field, p.flagName, 0, p.usage)
	}
	getopt.Parse()

	for _, p := range params {
		if envVal := os.Getenv(p.envName); envVal != "" {
			*p.field = envVal
		}
	}
}
