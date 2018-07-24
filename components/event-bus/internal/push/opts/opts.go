package opts

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/event-bus/internal/trace"
	"github.com/nats-io/go-nats-streaming"
)

const (
	defaultPort                   = 8080
	defaultClientID               = "event-bus-push"
	defaultNatsURL                = stan.DefaultNatsURL
	defaultConnectWait            = stan.DefaultConnectWait
	defaultNatsStreamingClusterID = "test-cluster"
	defaultAckWait                = stan.DefaultAckWait
	defaultMaxIdleConns           = 2
	defaultIdleConnTimeout        = 30 * time.Second
	defaultQueueGroup             = "event-bus-push"
)

// version is the current version for the event-bus-push.
var version = os.Getenv("APP_VERSION")

var usageStr = `
Usage: event-bus-push [options]
HTTP Server Options:
        --port <int>              HTTP server port (default: 8080)

NATS Client Options:
        --client_id <string>      Client ID to identify the client on the server (default: event-bus-push) (valid characters: alphanumeric, '-', '_')
        --nats_url <string>       URL NATS client should use to connect to the NATS server (default: nats://localhost:4222)
        --connect_wait <duration> Timeout used for the connect operation (default: 2s)

NATS Streaming Client Options:
        --cluster_id <string>     NATS Streaming cluster ID to connect to (default: test-cluster)
        --ack_wait <duration>     How long the server should wait for an ACK before resending a message (default: 30s)
		--max_inflight <int>      Maximum number of inflight messages with outstanding ACKs the server can send (default: 1024)

HTTP Client Options:
        --max_idle_conns <int>                Maximum number of idle HTTP connections (default: 2)
        --idle_conn_timeout <duration>        Idle HTTP connection timeout (default: 30s)
        --tls_skip_verify <bool>              Skip verifying certificate chain for TLS/HTTPS connections, e.g. to accept self-signed certificates (default: false)

Common Options:
    -h, --help                                Show this message
    -v, --version                             Show version.
`

// print out usage instructions.
func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

// Options represents command line
type Options struct {
	Port                   int
	ClientID               string
	NatsURL                string
	NatsStreamingClusterID string
	ConnectWait            time.Duration
	QueueGroup             string
	SubscriptionNameHeader string
	TopicHeader            string
	AckWait                time.Duration
	MaxIdleConns           int
	IdleConnTimeout        time.Duration
	TLSSkipVerify          bool
	trace.Options
	CheckEventsActivation bool
	// TODO TLS
}

var DefaultOptions = Options{
	ClientID:               defaultClientID,
	NatsURL:                defaultNatsURL,
	NatsStreamingClusterID: defaultNatsStreamingClusterID,
	ConnectWait:            defaultConnectWait,
	AckWait:                defaultAckWait,
	MaxIdleConns:           defaultMaxIdleConns,
	IdleConnTimeout:        defaultIdleConnTimeout,
	TLSSkipVerify:          false,
	QueueGroup:             defaultQueueGroup,
	CheckEventsActivation:  false,
}

// ParseFlags parses command line flags
func ParseFlags() *Options {
	fs := flag.NewFlagSet("push", flag.ExitOnError)
	fs.Usage = usage

	opts, err := configureOptions(fs, os.Args[1:],
		func() {
			fmt.Printf("push version %s, ", version)
			os.Exit(0)
		},
		fs.Usage)
	if err != nil {
		log.Fatalf("failed to parse command line flags: %v", err.Error()+"\n"+usageStr)
	}

	return opts
}

func configureOptions(fs *flag.FlagSet, args []string, printVersion, printHelp func()) (*Options, error) {
	opts := &DefaultOptions

	var (
		showVersion bool
		showHelp    bool
		err         error
	)

	fs.BoolVar(&showHelp, "h", false, "show this message")
	fs.BoolVar(&showHelp, "help", false, "show this message")
	fs.BoolVar(&showVersion, "version", false, "print version information")
	fs.BoolVar(&showVersion, "v", false, "print version information")
	fs.IntVar(&opts.Port, "port", defaultPort, "HTTP server port")
	fs.StringVar(&opts.ClientID, "client_id", defaultClientID, "client ID to use")
	fs.StringVar(&opts.NatsURL, "nats_url", defaultNatsURL, "NATS Server URL to connect to")
	fs.StringVar(&opts.NatsStreamingClusterID, "cluster_id", defaultNatsStreamingClusterID, "NATS Streaming Cluster ID to connect to")
	fs.DurationVar(&opts.ConnectWait, "connect_wait", defaultConnectWait, "NATS Streaming client connection timeout")
	fs.DurationVar(&opts.AckWait, "ack_wait", defaultAckWait, "NATS Streaming ack timeout before resending")
	fs.StringVar(&opts.QueueGroup, "queue_group", defaultQueueGroup, "queue group name")
	fs.BoolVar(&opts.TLSSkipVerify, "tls_skip_verify", false, "Skip TLS certificate verification, allow insecure connection")
	fs.StringVar(&opts.SubscriptionNameHeader, "subscription_name_header", "", "Push Subscription name header")
	fs.StringVar(&opts.TopicHeader, "topic_header", "", "Topic header")
	fs.StringVar(&opts.APIURL, "trace_api_url", trace.DefaultTraceAPIURL, "Trace API URL")
	fs.StringVar(&opts.HostPort, "trace_host_port", trace.DefaultTraceHostPort, "Trace host port")
	fs.StringVar(&opts.ServiceName, "trace_service_name", trace.DefaultTraceServiceName, "Push service name")
	fs.StringVar(&opts.OperationName, "trace_operation_name", trace.DefaultTraceOperationName, "Push operation name")
	fs.BoolVar(&opts.Debug, "trace_debug", false, "Trace debug")
	fs.BoolVar(&opts.CheckEventsActivation, "check_events_activation", false, "Check Events Activation before starting subscription")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if showVersion {
		printVersion()
		return nil, nil
	}

	if showHelp {
		printHelp()
		return nil, nil
	}

	showVersion, showHelp, err = processCommandLineArgs(fs)
	if err != nil {
		return nil, err
	} else if showVersion {
		printVersion()
		return nil, nil
	} else if showHelp {
		printHelp()
		return nil, nil
	}

	flagSet := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		flagSet[f.Name] = true
	})

	// TODO validate that mandatory flags with defaults are not empty, they haven't been cleared through CLI

	// TODO signal processing

	return opts, nil
}

func processCommandLineArgs(cmd *flag.FlagSet) (showVersion bool, showHelp bool, err error) {
	if len(cmd.Args()) > 0 {
		arg := cmd.Args()[0]
		switch strings.ToLower(arg) {
		case "version":
			return true, false, nil
		case "help":
			return false, true, nil
		default:
			return false, false, fmt.Errorf("unrecognized command: %q", arg)
		}
	}

	return false, false, nil
}
