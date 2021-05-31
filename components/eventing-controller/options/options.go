package options

import (
	"flag"
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

const (
	// args
	argNameMaxReconnects   = "max-reconnects"
	argNameMetricsAddr     = "metrics-addr"
	argNameReconnectWait   = "reconnect-wait"
	argNameReconcilePeriod = "reconcile-period"

	// env
	envNameLogFormat = "APP_LOG_FORMAT"
	envNameLogLevel  = "APP_LOG_LEVEL"
)

// Options represents the controller options.
type Options struct {
	Args
	Env
}

// Args represents the controller command-line arguments.
type Args struct {
	MaxReconnects   int
	MetricsAddr     string
	ReconnectWait   time.Duration
	ReconcilePeriod time.Duration
}

// Env represents the controller environment variables.
type Env struct {
	LogFormat string `envconfig:"APP_LOG_FORMAT" default:"json"`
	LogLevel  string `envconfig:"APP_LOG_LEVEL" default:"warn"`
}

// New returns a new Options instance.
func New() *Options {
	return &Options{}
}

// Parse parses the controller options.
func (o *Options) Parse() error {
	flag.IntVar(&o.MaxReconnects, argNameMaxReconnects, 10, "Maximum number of reconnect attempts (NATS).")
	flag.StringVar(&o.MetricsAddr, argNameMetricsAddr, ":8080", "The address the metric endpoint binds to.")
	flag.DurationVar(&o.ReconnectWait, argNameReconnectWait, time.Second, "Wait time between reconnect attempts (NATS).")
	flag.DurationVar(&o.ReconcilePeriod, argNameReconcilePeriod, time.Minute*10, "Period between triggering of reconciling calls (BEB).")
	flag.Parse()

	if err := envconfig.Process("", &o.Env); err != nil {
		return err
	}

	return nil
}

// String implements the fmt.Stringer interface.
func (o Options) String() string {
	return fmt.Sprintf("--%s=%v --%s=%v --%s=%v --%s=%v %s=%v %s=%v",
		argNameMaxReconnects, o.MaxReconnects,
		argNameMetricsAddr, o.MetricsAddr,
		argNameReconnectWait, o.ReconnectWait,
		argNameReconcilePeriod, o.ReconcilePeriod,
		envNameLogFormat, o.LogFormat,
		envNameLogLevel, o.LogLevel,
	)
}
