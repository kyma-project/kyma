package options

import "flag"

type Options struct {
	MaxRequestSize int64
	MetricsAddress string
}

func ParseArgs() *Options {
	maxRequestSize := flag.Int64("max-request-size", 65536, "The maximum request size in bytes.")
	metricsAddress := flag.String("metrics-addr", ":9090", "The address the metric endpoint binds to.")

	flag.Parse()

	return &Options{
		MaxRequestSize: *maxRequestSize,
		MetricsAddress: *metricsAddress,
	}
}
