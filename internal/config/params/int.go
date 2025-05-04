package params

import (
	"os"
	"strconv"

	"github.com/pborman/getopt/v2"
)

type Int struct {
	Field      *int
	FlagName   string
	EnvName    string
	DefaultVal int
	Usage      string
}

func (p *Int) SetDefault() {
	*p.Field = p.DefaultVal
}

func (p *Int) RegisterFlag() {
	getopt.FlagLong(p.Field, p.FlagName, 0, p.Usage)
}

func (p *Int) LoadFromEnv() {
	if envVal := os.Getenv(p.EnvName); envVal != "" {
		if val, err := strconv.Atoi(envVal); err == nil {
			*p.Field = val
		}
	}
}
