package util

import (
	"flag"
	"log"
	"os"
)

var version = os.Getenv("APP_VERSION")

const (
	dashboardURL   = "http://localhost:30000"
	defaultPort    = ":8080"
	secretFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

//Options ...
type Options struct {
	DashboardURL   string
	Port           string
	SecretFilePath string
}

//ParseFlags ...
func ParseFlags() *Options {
	fs := flag.NewFlagSet("reverseproxy", flag.ExitOnError)
	opts, err := configureOptions(fs, os.Args[1:])
	if err != nil {
		log.Fatalf("failed to parse command line flags: %v", err.Error())
	}
	return opts
}

func configureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	opts := defaultOptions()
	var showHelp bool

	fs.BoolVar(&showHelp, "show-help", false, "Print the command line options")
	fs.StringVar(&opts.Port, "port", defaultPort, "The reverse proxy listen port")
	fs.StringVar(&opts.SecretFilePath, "secret-token-path", secretFilePath, "Secret token file path")
	fs.StringVar(&opts.DashboardURL, "k8s-dashboard-URL", dashboardURL, "Kubernetes Dashboard URL")

	if err := fs.Parse(args); err != nil {
		log.Printf("a77a!")
		return nil, err
	}

	if showHelp {
		fs.Usage()
		flag.PrintDefaults()
		os.Exit(0)
	}
	return opts, nil
}

func defaultOptions() *Options {
	return &Options{
		Port:           defaultPort,
		SecretFilePath: secretFilePath,
		DashboardURL:   dashboardURL,
	}
}
