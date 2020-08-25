package tracing

import (
	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkingo "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/pkg/errors"
	"strconv"
)

func Setup(cfg Config, hostPort int) error {
	// set up a span reporter
	reporter := zipkinhttp.NewReporter(cfg.CollectorUrl)
	defer reporter.Close()

	// create our local service endpoint
	endpoint, err := zipkingo.NewEndpoint(cfg.ServiceSpanName, strconv.Itoa(hostPort))
	if err != nil {
		return errors.Wrap(err, " while creating local endpoint")
	}

	// initialize our tracer
	nativeTracer, err := zipkingo.NewTracer(reporter, zipkingo.WithLocalEndpoint(endpoint))
	if err != nil {
		return errors.Wrap(err, " while creating tracer")
	}

	// use zipkin-go-opentracing to wrap our tracer
	tracer := zipkinot.Wrap(nativeTracer)

	// optionally set as Global OpenTracing tracer instance
	opentracing.SetGlobalTracer(tracer)
	return nil
}
