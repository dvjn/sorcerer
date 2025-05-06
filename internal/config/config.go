package config

import (
	"os"

	"github.com/dvjn/sorcerer/internal/config/params"
	"github.com/pborman/getopt/v2"
)

type ProxyHeaderAuth struct {
	UserHeaderName        string
	GroupsHeaderName      string
	GroupsHeaderSeparator string
}

type AuthConfig struct {
	Mode            string
	ProxyHeaderAuth ProxyHeaderAuth
}

type Config struct {
	StoragePath string
	Port        int
	Auth        AuthConfig
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
		&params.String{
			Field:      &config.Auth.Mode,
			FlagName:   "auth-mode",
			EnvName:    "AUTH_MODE",
			DefaultVal: "",
			Usage:      "Should be one of none, proxy-header",
		},
		&params.String{
			Field:      &config.Auth.ProxyHeaderAuth.UserHeaderName,
			FlagName:   "auth-user-header",
			EnvName:    "AUTH_USER_HEADER",
			DefaultVal: "",
			Usage:      "The header to use for fetching the user name",
		},
		&params.String{
			Field:      &config.Auth.ProxyHeaderAuth.GroupsHeaderName,
			FlagName:   "auth-groups-header",
			EnvName:    "AUTH_GROUPS_HEADER",
			DefaultVal: "",
			Usage:      "The header to use for fetching the user groups",
		},
		&params.String{
			Field:      &config.Auth.ProxyHeaderAuth.GroupsHeaderSeparator,
			FlagName:   "auth-groups-header-sep",
			EnvName:    "AUTH_GROUPS_HEADER_SEP",
			DefaultVal: ",",
			Usage:      "The separator to use for the groups",
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
