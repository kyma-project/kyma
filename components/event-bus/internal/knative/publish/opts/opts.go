package opts

import (
	"flag"
	"log"
	"strings"

	"github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/kyma-project/kyma/components/event-bus/internal/trace"
)

const (
	// default option values
	defaultPort                 = 8080
	defaultMaxRequests          = 100
	defaultMaxRequestSize       = 65536
	defaultMaxChannelNameLength = 45

	// option names
	port                      = "port"
	maxRequests               = "max_requests"
	maxRequestSize            = "max_request_size"
	maxChannelNameLength      = "max_channel_name_length"
	traceDebug                = "trace_debug"
	traceAPIURL               = "trace_api_url"
	traceHostPort             = "trace_host_port"
	traceServiceName          = "trace_service_name"
	traceOperationName        = "trace_operation_name"
	maxSourceIDLength         = "max_source_id_length"
	maxEventTypeLength        = "max_event_type_length"
	maxEventTypeVersionLength = "max_event_type_version_length"
)

type Options struct {
	Port                 int
	MaxRequests          int
	MaxRequestSize       int64
	MaxChannelNameLength int
	TraceOptions         *trace.Options
	EventOptions         *publish.EventOptions
}

func ParseFlags() *Options {
	opts := GetDefaultOptions()

	// application opts
	flag.IntVar(&opts.Port, port, defaultPort, "The port used to communicate with the Knative Publish service endpoints")
	flag.IntVar(&opts.MaxRequests, maxRequests, defaultMaxRequests, "The maximum number of allowed concurrent requests handled by the Knative Publish service")
	flag.Int64Var(&opts.MaxRequestSize, maxRequestSize, defaultMaxRequestSize, "The maximum request size in bytes")
	flag.IntVar(&opts.MaxChannelNameLength, maxChannelNameLength, defaultMaxChannelNameLength, "The maximum channel name length")
	// trace opts
	flag.BoolVar(&opts.TraceOptions.Debug, traceDebug, trace.DefaultTraceDebug, "The Trace debug flag")
	flag.StringVar(&opts.TraceOptions.APIURL, traceAPIURL, trace.DefaultTraceAPIURL, "The Trace API URL")
	flag.StringVar(&opts.TraceOptions.HostPort, traceHostPort, trace.DefaultTraceHostPort, "The Trace host port")
	flag.StringVar(&opts.TraceOptions.ServiceName, traceServiceName, trace.DefaultTraceServiceName, "The Knative Publish service name")
	flag.StringVar(&opts.TraceOptions.OperationName, traceOperationName, trace.DefaultTraceOperationName, "The Knative Publish operation name")
	// event opts
	flag.IntVar(&opts.EventOptions.MaxSourceIDLength, maxSourceIDLength, publish.DefaultMaxSourceIDLength, "The maximum source id length")
	flag.IntVar(&opts.EventOptions.MaxEventTypeLength, maxEventTypeLength, publish.DefaultMaxEventTypeLength, "The maximum event type length")
	flag.IntVar(&opts.EventOptions.MaxEventTypeVersionLength, maxEventTypeVersionLength, publish.DefaultMaxEventTypeVersionLength, "The maximum event type version length")

	flag.Parse()
	return opts
}

func GetDefaultOptions() *Options {
	opts := &Options{
		Port:                 defaultPort,
		MaxRequests:          defaultMaxRequests,
		MaxRequestSize:       defaultMaxRequestSize,
		MaxChannelNameLength: defaultMaxChannelNameLength,
		TraceOptions:         trace.GetDefaultTraceOptions(),
		EventOptions:         publish.GetDefaultEventOptions(),
	}
	return opts
}

func (options *Options) Print() {
	log.Println(strings.Repeat("-", 50))
	log.Printf(" %s %v", port, options.Port)
	log.Printf(" %s %v", maxRequests, options.MaxRequests)
	log.Printf(" %s %v", maxRequestSize, options.MaxRequestSize)
	log.Printf(" %s %v", maxChannelNameLength, options.MaxChannelNameLength)
	log.Printf(" %s %v", traceDebug, options.TraceOptions.Debug)
	log.Printf(" %s %v", traceAPIURL, options.TraceOptions.APIURL)
	log.Printf(" %s %v", traceHostPort, options.TraceOptions.HostPort)
	log.Printf(" %s %v", traceServiceName, options.TraceOptions.ServiceName)
	log.Printf(" %s %v", traceOperationName, options.TraceOptions.OperationName)
	log.Printf(" %s %v", maxSourceIDLength, options.EventOptions.MaxSourceIDLength)
	log.Printf(" %s %v", maxEventTypeLength, options.EventOptions.MaxEventTypeLength)
	log.Printf(" %s %v", maxEventTypeVersionLength, options.EventOptions.MaxEventTypeVersionLength)
	log.Println(strings.Repeat("-", 50))
}
