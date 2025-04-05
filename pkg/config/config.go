package config

import (
	"os"

	"github.com/dvjn/sorcerer/pkg/config/params"
	"github.com/pborman/getopt/v2"
)

type Config struct {
	StoragePath string
	Port        int
}

type param interface {
	SetDefault()
	RegisterFlag()
	LoadFromEnv()
}

func LoadConfig() *Config {
	var config Config

	params := []param{
		&params.String{
			Field:      &config.StoragePath,
			FlagName:   "storage-path",
			EnvName:    "STORAGE_PATH",
			DefaultVal: "data",
			Usage:      "registry data storage path",
		},
		&params.Int{
			Field:      &config.Port,
			FlagName:   "port",
			EnvName:    "PORT",
			DefaultVal: 3000,
			Usage:      "The port to run the server on",
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
		p.SetDefault()
	}

	for _, p := range params {
		p.RegisterFlag()
	}
	getopt.Parse()

	for _, p := range params {
		p.LoadFromEnv()
	}
}
