package opts

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"
)

const (
	defaultPort           = 8080
	defaultResyncPeriod   = 10 * time.Second
	defaultChannelTimeout = 10 * time.Second
)

// Options represents the subscription options.
type Options struct {
	Port           int
	ResyncPeriod   time.Duration
	ChannelTimeout time.Duration
}

var (
	config *Options
	once   sync.Once
)

// GetOptions returns an instance of the subscription options.
func GetOptions() *Options {
	once.Do(func() {
		config = ParseFlags()
	})
	return config
}

// ParseFlags parses the command line flags.
func ParseFlags() *Options {
	fs := flag.NewFlagSet("sv", flag.ExitOnError)
	opts, err := configureOptions(fs, os.Args[1:])

	if err != nil {
		log.Fatalf("failed to parse command line flags: %v", err.Error())
	}

	config = opts
	return opts
}

func configureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	opts := DefaultOptions()
	var showHelp bool

	fs.IntVar(&opts.Port, "port", defaultPort, "The subscription controller knative healtcheck listen port")
	fs.DurationVar(&opts.ResyncPeriod, "resyncPeriod", defaultResyncPeriod, "The resync period for the used informers")
	fs.DurationVar(&opts.ChannelTimeout, "channelTimeout", defaultChannelTimeout, "The timeout for Knative Channel creation")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if showHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}
	return opts, nil
}

// DefaultOptions returns the default subscription options.
func DefaultOptions() *Options {
	return &Options{
		Port:           defaultPort,
		ResyncPeriod:   defaultResyncPeriod,
		ChannelTimeout: defaultChannelTimeout,
	}
}
