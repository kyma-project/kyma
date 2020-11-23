package options

import "flag"

type Options struct {
	MaxRequestSize int64
}

func ParseArgs() *Options {
	maxRequestSize := flag.Int64("maxRequestSize", 65536, "The maximum request size in bytes")
	flag.Parse()

	return &Options{
		MaxRequestSize: *maxRequestSize,
	}
}
