package main

import (
	"time"

	"github.com/caarlos0/env"
)

type config struct {
	CASBucket string        `env:"CAS_BUCKET,required"`
	CASPrefix string        `env:"CAS_PREFIX"`
	ACBucket  string        `env:"AC_BUCKET,required"`
	ACPrefix  string        `env:"AC_PREFIX"`
	Timeout   time.Duration `env:"S3_TIMEOUT"`
	Listen    string        `env:"LISTEN" envDefault:":http"`
	LogLevel  string        `env:"LOG_LEVEL" envDefault:"info"`
}

func parseConfig() (*config, error) {
	cfg := &config{}
	err := env.Parse(cfg)
	return cfg, err
}
