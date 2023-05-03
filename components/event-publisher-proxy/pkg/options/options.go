package options

import (
	"flag"
	"fmt"
)

const (
	maxRequestSize     = 65536
	metricEndpointPort = ":9090"

	// All the available arguments.
	argMaxRequestSize = "max-request-size"
	argMetricsAddress = "metrics-addr"
)

type Options struct {
	MaxRequestSize int64
	MetricsAddress string
}

func New() *Options {
	return &Options{}
}

func (o *Options) Parse() error {
	flag.Int64Var(&o.MaxRequestSize, argMaxRequestSize, maxRequestSize, "The maximum request size in bytes.")
	flag.StringVar(&o.MetricsAddress, argMetricsAddress, metricEndpointPort, "The address the metric endpoint binds to.")
	flag.Parse()

	return nil
}

func (o Options) String() string {
	return fmt.Sprintf("--%s=%v --%s=%v",
		argMaxRequestSize, o.MaxRequestSize,
		argMetricsAddress, o.MetricsAddress,
	)
}
