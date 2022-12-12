package options

import (
	"flag"
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

const (
	// All the available arguments.
	argMaxRequestSize = "max-request-size"
	argMetricsAddress = "metrics-addr"
	// All the available environment variables.
	envEnableNewCRDVersion = "ENABLE_NEW_CRD_VERSION"
)

type Options struct {
	MaxRequestSize int64
	MetricsAddress string
	Env
}

// Env represents the controller environment variables.
type Env struct {
	EnableNewCRDVersion bool `envconfig:"ENABLE_NEW_CRD_VERSION" default:"false"`
}

func New() *Options {
	return &Options{}
}

func (o *Options) Parse() error {
	flag.Int64Var(&o.MaxRequestSize, argMaxRequestSize, 65536, "The maximum request size in bytes.")
	flag.StringVar(&o.MetricsAddress, argMetricsAddress, ":9090", "The address the metric endpoint binds to.")
	flag.Parse()

	if err := envconfig.Process("", &o.Env); err != nil {
		return err
	}
	return nil
}

func (o Options) String() string {
	return fmt.Sprintf("--%s=%v --%s=%v %s=%v",
		argMaxRequestSize, o.MaxRequestSize,
		argMetricsAddress, o.MetricsAddress,
		envEnableNewCRDVersion, o.EnableNewCRDVersion,
	)
}
