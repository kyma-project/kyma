package opts

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"
)

var version = os.Getenv("APP_VERSION")

const (
	defaultPort         = 8080
	defaultResyncPeriod = 1 * time.Minute
)

type Options struct {
	Port         int
	ResyncPeriod time.Duration
}

var (
	config *Options
	once   sync.Once
)

func GetOptions() *Options {
	once.Do(func() {
		config = ParseFlags()
	})
	return config
}

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

	fs.IntVar(&opts.Port, "port", defaultPort, "The publish listen port")
	fs.DurationVar(&opts.ResyncPeriod, "resyncPeriod", defaultResyncPeriod, "The resync period for the used informers")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if showHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}
	return opts, nil
}

func DefaultOptions() *Options {
	return &Options{
		Port:         defaultPort,
		ResyncPeriod: defaultResyncPeriod,
	}
}
