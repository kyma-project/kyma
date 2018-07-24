package publish

import (
	"flag"
	"log"
	"os"

	"github.com/kyma-project/kyma/components/event-bus/internal/trace"
	"github.com/nats-io/go-nats-streaming"
)

var version = os.Getenv("APP_VERSION")

const (
	defaultPort                   = 8080
	defaultNatsURL                = stan.DefaultNatsURL
	defaultClientID               = "kyma-publish"
	defaultMaxRequests            = 100
	defaultNatsStreamingClusterID = "kyma-nats-streaming"
	defaultTraceAPIURL            = "http://localhost:9411/api/v1/spans"
	defaultTraceHostPort          = "0.0.0.0:0"
	defaultServiceName            = "publish-service"
	defaultOperationName          = "publish-to-NATS"
)

type Options struct {
	Port                   int
	NatsURL                string
	ClientID               string
	NoOfConcurrentRequests int
	NatsStreamingClusterID string
	TraceRequests          bool
	TraceAPIURL            string
	TraceHostPort          string
	ServiceName            string
	TraceDebug             bool
	OperationName          string
}

func ParseFlags() *Options {
	fs := flag.NewFlagSet("publish", flag.ExitOnError)
	opts, err := configureOptions(fs, os.Args[1:])

	if err != nil {
		log.Fatalf("failed to parse command line flags: %v", err.Error())
	}

	return opts
}

func configureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	opts := DefaultOptions()
	var showHelp bool

	fs.IntVar(&opts.Port, "port", defaultPort, "The publish listen port")
	fs.StringVar(&opts.NatsURL, "nats_url", defaultNatsURL, "The NATS URL")
	fs.StringVar(&opts.ClientID, "client_id", defaultClientID, "client ID to use")
	fs.IntVar(&opts.NoOfConcurrentRequests, "max_requests", defaultMaxRequests, "The max number of accepted concurrent requests")
	fs.StringVar(&opts.NatsStreamingClusterID, "nats_streaming_cluster_id", defaultNatsStreamingClusterID, "The NATS Streaming cluster id")
	fs.BoolVar(&opts.TraceRequests, "trace", false, "Log verbosily the received HTTP requests traces.")
	fs.BoolVar(&showHelp, "showHelp", false, "Print the command line options")
	fs.StringVar(&opts.TraceAPIURL, "trace_api_url", defaultTraceAPIURL, "Trace API URL")
	fs.StringVar(&opts.TraceHostPort, "trace_host_port", defaultTraceHostPort, "Trace host port")
	fs.StringVar(&opts.ServiceName, "service_name", defaultServiceName, "Publish service name")
	fs.StringVar(&opts.OperationName, "operation_name", defaultOperationName, "Publish operation name")
	fs.BoolVar(&opts.TraceDebug, "trace_debug", false, "Trace debug")
	fs.StringVar(&opts.OperationName, "trace_operation_name", trace.DefaultTraceOperationName, "Publish operation name")
	fs.StringVar(&opts.ServiceName, "trace_service_name", trace.DefaultTraceServiceName, "Publish service name")

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
		Port:                   defaultPort,
		ClientID:               defaultClientID,
		NoOfConcurrentRequests: defaultMaxRequests,
		OperationName:          defaultOperationName,
		ServiceName:            defaultServiceName,
		TraceHostPort:          defaultTraceHostPort,
		TraceAPIURL:            defaultTraceAPIURL,
		NatsStreamingClusterID: defaultNatsStreamingClusterID,
		NatsURL:                defaultNatsURL,
	}
}
