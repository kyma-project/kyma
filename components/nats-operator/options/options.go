package options

import (
	"flag"
	"time"
)

const (
	argNameInterval = "interval"
)

// Options represents the application options.
type Options struct {
	Interval time.Duration
}

// New returns a new Options instance.
func New() *Options {
	return &Options{}
}

// Parse parses the application options.
func (o *Options) Parse() *Options {
	flag.DurationVar(&o.Interval, argNameInterval, time.Minute, "The duration between consequent health-checks.")
	flag.Parse()
	return o
}
