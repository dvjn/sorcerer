package params

import (
	"os"

	"github.com/pborman/getopt/v2"
)

type String struct {
	Field      *string
	FlagName   string
	EnvName    string
	DefaultVal string
	Usage      string
}

func (p *String) SetDefault() {
	*p.Field = p.DefaultVal
}

func (p *String) RegisterFlag() {
	getopt.FlagLong(p.Field, p.FlagName, 0, p.Usage)
}

func (p *String) LoadFromEnv() {
	if envVal := os.Getenv(p.EnvName); envVal != "" {
		*p.Field = envVal
	}
}
