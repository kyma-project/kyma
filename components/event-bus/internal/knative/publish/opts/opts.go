package opts

import (
	"flag"
	"log"
	"strings"

	"github.com/kyma-project/kyma/components/event-bus/internal/trace"
)

const (
	defaultPort                 = 8080
	defaultMaxRequests          = 100
	defaultMaxRequestSize       = 65536
	DefaultMaxChannelNameLength = 63 // max channel name length by knative
	defaultTraceDebug           = false
	defaultTraceAPIURL          = "http://localhost:9411/api/v1/spans"
	defaultTraceHostPort        = "0.0.0.0:0"
	defaultServiceName          = "knative-publish-service"
	defaultOperationName        = "publish-to-knative"
)

type Options struct {
	Port                 int
	MaxRequests          int
	MaxRequestSize       int64
	MaxChannelNameLength int
	TraceOptions         *trace.Options
}

func ParseFlags() *Options {
	opts := &Options{TraceOptions: &trace.Options{}}
	flag.IntVar(&opts.Port, "port", defaultPort, "The port used to communicate with the Knative Publish service endpoints")
	flag.IntVar(&opts.MaxRequests, "max_requests", defaultMaxRequests, "The maximum number of allowed concurrent requests handled by the Knative Publish service")
	flag.Int64Var(&opts.MaxRequestSize, "max_request_size", defaultMaxRequestSize, "The maximum request size in bytes")
	flag.BoolVar(&opts.TraceOptions.Debug, "trace_debug", defaultTraceDebug, "The Trace debug flag")
	flag.IntVar(&opts.MaxChannelNameLength, "max_channel_name_length", DefaultMaxChannelNameLength, "The maximum channel name length")
	flag.StringVar(&opts.TraceOptions.APIURL, "trace_api_url", defaultTraceAPIURL, "The Trace API URL")
	flag.StringVar(&opts.TraceOptions.HostPort, "trace_host_port", defaultTraceHostPort, "The Trace host port")
	flag.StringVar(&opts.TraceOptions.ServiceName, "trace_service_name", defaultServiceName, "The Knative Publish service name")
	flag.StringVar(&opts.TraceOptions.OperationName, "trace_operation_name", defaultOperationName, "The Knative Publish operation name")
	flag.Parse()
	return opts
}

func DefaultOptions() *Options {
	opts := &Options{
		Port:                 defaultPort,
		MaxRequests:          defaultMaxRequests,
		MaxRequestSize:       defaultMaxRequestSize,
		MaxChannelNameLength: DefaultMaxChannelNameLength,
		TraceOptions: &trace.Options{
			Debug:         defaultTraceDebug,
			APIURL:        defaultTraceAPIURL,
			HostPort:      defaultTraceHostPort,
			ServiceName:   defaultServiceName,
			OperationName: defaultOperationName,
		},
	}
	return opts
}

func (options *Options) Print() {
	log.Println(strings.Repeat("-", 50))
	log.Printf(" port %v", options.Port)
	log.Printf(" max_requests %v", options.MaxRequests)
	log.Printf(" max_request_size %v", options.MaxRequestSize)
	log.Printf(" max_channel_name_length %v", options.MaxChannelNameLength)
	log.Printf(" trace_debug %v", options.TraceOptions.Debug)
	log.Printf(" trace_api_url %v", options.TraceOptions.APIURL)
	log.Printf(" trace_host_port %v", options.TraceOptions.HostPort)
	log.Printf(" trace_service_name %v", options.TraceOptions.ServiceName)
	log.Printf(" trace_operation_name %v", options.TraceOptions.OperationName)
	log.Println(strings.Repeat("-", 50))
}
